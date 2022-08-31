package tasks

import (
	"github.com/headwaymaps/headway/cmd/headway-build/tasks/task"
)

type RunContainerOptions struct {
	Command []string
	Volumes []ContainerVolume
}

type ContainerVolume struct {
	Destination string
	Source      string
	Options     []string
}

func RunContainer(image string, options RunContainerOptions) {
	// TODO: detect podman or docker
	task.Execute(&Podman3RunTask{
		image: image,
		opts:  options,
	})
}
