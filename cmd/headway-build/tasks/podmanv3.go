package tasks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/containers/podman/v3/libpod/define"
	"github.com/containers/podman/v3/pkg/bindings"
	"github.com/containers/podman/v3/pkg/bindings/containers"
	"github.com/containers/podman/v3/pkg/bindings/images"
	"github.com/containers/podman/v3/pkg/specgen"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/headwaymaps/headway/cmd/headway-build/tasks/task"
)

type Podman3RunTask struct {
	image   string
	started *time.Time
	status  string
	opts    RunContainerOptions
}

func (p Podman3RunTask) View() string {
	if p.started == nil {
		return p.status
	}

	return fmt.Sprintf("%s (%s)", p.status, time.Since(*p.started).Truncate(time.Second))
}

func (p *Podman3RunTask) Run() (task.Result, error) {
	p.restartTimer("connecting to podman")

	ctx, err := bindings.NewConnection(context.Background(), getPodmanSocket())
	if err != nil {
		return task.Result{}, fmt.Errorf("error connecting to podman: %v", err)
	}

	p.restartTimer(fmt.Sprintf("pulling image %s", p.image))

	t := true
	_, err = images.Pull(ctx, p.image, &images.PullOptions{Quiet: &t})
	if err != nil {
		return task.Result{}, fmt.Errorf("error pulling image %s: %v", p.image, err)
	}

	p.restartTimer(fmt.Sprintf("booting container"))

	s := specgen.NewSpecGenerator(p.image, false)
	s.Command = p.opts.Command
	s.Mounts = []specs.Mount{}

	for _, vol := range p.opts.Volumes {
		mount, err := getPodman3Mount(vol)
		if err != nil {
			return task.Result{}, fmt.Errorf("invalid volume mount request: %v", err)
		}

		s.Mounts = append(s.Mounts, mount)
	}

	r, err := containers.CreateWithSpec(ctx, s, &containers.CreateOptions{})
	if err != nil {
		return task.Result{}, fmt.Errorf("error creating container: %v", err)
	}

	for _, warning := range r.Warnings {
		fmt.Println("⚠️ " + warning)
	}

	err = containers.Start(ctx, r.ID, &containers.StartOptions{})
	if err != nil {
		return task.Result{}, fmt.Errorf("error starting container: %s", err)
	}

	_, err = containers.Wait(ctx, r.ID, &containers.WaitOptions{
		Condition: []define.ContainerStatus{define.ContainerStateRunning},
	})
	if err != nil {
		return task.Result{}, fmt.Errorf("error waiting for container to start: %v", err)
	}

	p.restartTimer(fmt.Sprintf("running %s", p.image))

	_, err = containers.Wait(ctx, r.ID, &containers.WaitOptions{
		Condition: []define.ContainerStatus{define.ContainerStateStopped},
	})
	if err != nil {
		return task.Result{}, fmt.Errorf("error waiting for container to start: %v", err)
	}

	p.status = fmt.Sprintf("ran %s", p.image)

	return task.Result{}, nil
}

func (p *Podman3RunTask) restartTimer(status string) {
	t := time.Now()
	p.started = &t
	p.status = status
}

func getPodmanSocket() string {
	sockDir := os.Getenv("XDG_RUNTIME_DIR")
	if sockDir == "" {
		sockDir = "/var/run"
	}
	return fmt.Sprintf("unix:%s/podman/podman.sock", sockDir)
}

func getPodman3Mount(vol ContainerVolume) (specs.Mount, error) {
	source, err := filepath.Abs(vol.Source)
	if err != nil {
		return specs.Mount{}, err
	}

	return specs.Mount{
		Destination: vol.Destination,
		Source: source,
		Options: vol.Options,
		Type: "bind",
	}, nil
}