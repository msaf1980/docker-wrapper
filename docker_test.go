package docker_wrapper

import (
	"errors"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var pid string = strconv.Itoa(os.Getpid())

func TestContainer(t *testing.T) {
	tests := []struct {
		name    string
		docker  string
		image   string
		tag     string
		exposes []string
		volumes []string
		links   []string
		limits  []string
		envs    []string
		wantErr bool
	}{
		{
			image: "redis",
			tag:   "6.2",
			name:  "redis-6.2-" + pid,
		},
	}
	for i, tt := range tests {
		t.Run(tt.name+"#"+strconv.Itoa(i), func(t *testing.T) {
			c := &Container{
				Docker:  tt.docker,
				Image:   tt.image,
				Tag:     tt.tag,
				Name:    tt.name,
				Exposes: tt.exposes,
				Volumes: tt.volumes,
				Links:   tt.links,
				Limits:  tt.limits,
				Envs:    tt.envs,
			}
			err := c.Start()
			if (err != nil) != tt.wantErr {
				t.Errorf("Container.Start() error = '%v', wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				assert.True(t, c.Id() == "")
			} else {
				assert.True(t, c.Id() != "")

				err = c.IsRunning()
				assert.NoErrorf(t, err, "Container.IsRunning()")

				// stop and delete
				err = c.Stop(true)
				assert.NoErrorf(t, err, "Container.Stop()")

				err = c.IsRunning()
				assert.True(t, errors.Is(err, ErrIdNotSet), "double Container.Delete() = %v", err)
			}
		})
	}
}

func TestContainer_IsRunning(t *testing.T) {
	c := &Container{
		Image: "redis",
		Tag:   "6.2",
		Name:  "redis-6.2" + pid,
	}
	err := c.Start()
	require.NoErrorf(t, err, "Container.Start() %s")

	assert.True(t, c.Id() != "")

	err = c.IsRunning()
	assert.NoErrorf(t, err, "Container.IsRunning()")

	// stop
	err = c.Stop(false)
	assert.NoErrorf(t, err, "Container.Stop() %s")

	err = c.IsRunning()
	assert.Truef(t, errors.Is(err, ErrNotRunning), "Container.IsRunning() = %v", err)

	// double stop
	err = c.Stop(false)
	assert.NoErrorf(t, err, "Container.Stop() %s")

	id := c.Id()

	err = c.IsStopped()
	assert.NoErrorf(t, err, "Container.IsStopped() %s")

	err = c.IsExist()
	assert.NoErrorf(t, err, "Container.IsExist() %s")

	// double stop and first delete
	err = c.Stop(true)
	assert.NoErrorf(t, err, "Container.Stop() %s")

	err = c.IsRunning()
	assert.Truef(t, errors.Is(err, ErrIdNotSet), "Container.IsRunning() = %v", err)

	// double delete
	err = c.Delete()
	assert.Truef(t, errors.Is(err, ErrIdNotSet), "Container.Delete() = %v", err)

	c.Attach(id)
	err = c.IsRunning()
	assert.Truef(t, errors.Is(err, ErrNotExist), "Container.IsRunning() = %v", err)

	err = c.IsExist()
	assert.Truef(t, errors.Is(err, ErrNotExist), "Container.IsExist() = %v", err)
}
