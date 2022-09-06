package containers

import (
	"github.com/headwaymaps/headway/cmd/headway-build/tasks/task"
)

// Options for running a container
type Options struct {
	Image      string
	Name       TaskName
	Entrypoint []string
	Command    []string
	Volumes    []Volume
	User       string
}

// Volume to mount to a container
type Volume struct {
	Destination string
	Source      string
	Options     []string
}

// TaskName is the name of the task to display before, during and after the container runs
type TaskName struct {
	Before string // Before is printed to the screen followed by the suffix before the task has started
	During string // During is printed to the screen followed by the suffix while the task is running
	After  string // After is printed to the screen followed by the suffix after the task has run
	Suffix string // Suffix appears after the prefix for all states of the task
}

// RunContainer starts a container and waits for it to exit
func RunContainer(options Options) {
	// TODO: detect podman or docker
	task.Execute(&podman3RunTask{opts: options})
}
