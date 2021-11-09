package docker_wrapper

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var (
	// Docker container errors
	ErrIdNotSet = errors.New("id not set")
	ErrStarted  = errors.New("container is started")

	containerNotExist = "container not exist"
	ErrNotExist       = errors.New(containerNotExist)

	containerNotRunning = "container not running"
	ErrNotRunning       = errors.New(containerNotRunning)

	ErrNameIsInUse = errors.New("container with name is already in use")
	ErrParse       = errors.New("parse output error")
	ErrImageNotSet = errors.New("image not set")
)

type DockerError struct {
	Err error
	Out string
}

func (e *DockerError) Unwrap() error { return e.Err }

func (e *DockerError) Error() string {
	if len(e.Out) == 0 {
		return e.Err.Error()
	}
	return e.Err.Error() + ": " + e.Out
}

var (
	reContainerIsNotRunning = regexp.MustCompile(" Container .* is not running")
	reContainerIsInUse      = regexp.MustCompile(" Container .* is already in use")
	reContainerNotExist     = regexp.MustCompile(" No such container")
)

func dockerError(out string, origErr error) error {
	if origErr == nil {
		return nil
	}
	if reContainerIsNotRunning.MatchString(out) {
		return &DockerError{Err: ErrNotRunning, Out: out}
	}
	if reContainerIsInUse.MatchString(out) {
		return &DockerError{Err: ErrNameIsInUse, Out: out}
	}
	if reContainerNotExist.MatchString(out) {
		return &DockerError{Err: ErrNotExist, Out: out}
	}
	return &DockerError{Err: origErr, Out: out}
}

// Container wrapper for docker container
type Container struct {
	Docker string // Docker binary (default docker)

	Image string // Image name
	Tag   string // Tag

	Name string // Name of the container (optional)

	Exposes []string // Exposed ports (optional)
	Volumes []string // Volumes (optional)
	Links   []string // Links (optional)
	Limits  []string //	Limits (optional) (like nofile=262144:262144)
	Envs    []string // Envirment variables (optional)

	containerID string
}

// Container.Start start container
func (c *Container) Start() error {
	if len(c.Docker) == 0 {
		c.Docker = "docker"
	}

	if len(c.Image) == 0 {
		return fmt.Errorf("image not set")
	}
	if len(c.Tag) == 0 {
		c.Tag = "latest"
	}

	if len(c.containerID) > 0 {
		return ErrStarted
	}

	opts := make([]string, 0, 5+len(c.Exposes)+len(c.Volumes)+len(c.Links)+len(c.Limits))

	opts = append(opts, "run", "-d") // run dettached
	if len(c.Name) > 0 {
		opts = append(opts, "--name", c.Name)
	}
	for _, expose := range c.Exposes {
		opts = append(opts, "-p", expose)
	}
	for _, vol := range c.Volumes {
		opts = append(opts, "-v", vol)
	}
	for _, link := range c.Links {
		opts = append(opts, "--link", link)
	}
	for _, limit := range c.Limits {
		opts = append(opts, "--ulimit", limit)
	}
	opts = append(opts, c.Image+":"+c.Tag)

	cmd := exec.Command(c.Docker, opts...)
	out, err := cmd.CombinedOutput()
	outStr := string(out)
	if err == nil {
		s := strings.Split(outStr, "\n")
		if len(s) == 2 && len(s[1]) == 0 {
			c.containerID = s[0]
		} else {
			err = dockerError(outStr, ErrParse)
		}
	} else {
		err = dockerError(outStr, err)
	}

	return err
}

// Container.Stop stop container
// @param delete delete container after stop is successed
func (c *Container) Stop(delete bool) error {
	if len(c.containerID) == 0 {
		return ErrIdNotSet
	}

	chStop := []string{"stop", c.containerID}

	cmd := exec.Command(c.Docker, chStop...)
	out, err := cmd.CombinedOutput()
	outStr := string(out)

	if err == nil && delete {
		return c.Delete()
	} else {
		err = dockerError(outStr, err)
	}
	return err
}

// Container.Delete delete container
func (c *Container) Delete() error {
	if len(c.containerID) == 0 {
		return ErrIdNotSet
	}

	chDel := []string{"rm", c.containerID}

	cmd := exec.Command(c.Docker, chDel...)
	out, err := cmd.CombinedOutput()
	outStr := string(out)

	if err == nil {
		c.containerID = ""
	} else {
		err = dockerError(outStr, err)
	}

	return err
}

// Container.Attach attach container by id
func (c *Container) Attach(id string) {
	c.containerID = id
}

// Container.Detach detach container
func (c *Container) Detach() {
	c.containerID = ""
}

// Container.Id container Id
func (c *Container) Id() string {
	return c.containerID
}

// Container.IsRunning container is running
func (c *Container) IsRunning() error {
	if len(c.containerID) == 0 {
		return ErrIdNotSet
	}

	cmd := exec.Command(c.Docker, "top", c.containerID)
	out, err := cmd.CombinedOutput()

	return dockerError(string(out), err)
}

// Container.IsStopped container is stopped
func (c *Container) IsStopped() error {
	if len(c.containerID) == 0 {
		return ErrIdNotSet
	}

	cmd := exec.Command(c.Docker, "top", c.containerID)
	out, err := cmd.CombinedOutput()

	err = dockerError(string(out), err)
	if err == nil {
		return &DockerError{Err: ErrStarted}
	} else if errors.Is(err, ErrNotRunning) {
		return nil
	}

	return err
}

// Container.IsExist container is exist
func (c *Container) IsExist() error {
	if len(c.containerID) == 0 {
		return ErrIdNotSet
	}

	cmd := exec.Command(c.Docker, "top", c.containerID)
	out, err := cmd.CombinedOutput()

	err = dockerError(string(out), err)
	if err == nil {
		return nil
	} else if errors.Is(err, ErrNotRunning) {
		return nil
	}

	return err
}
