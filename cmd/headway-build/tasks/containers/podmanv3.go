package containers

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

type podman3RunTask struct {
	started *time.Time
	status  string
	opts    Options
	log     containerLogger
	done    bool
}

func (p podman3RunTask) View() string {
	if p.started == nil {
		return p.status
	}

	out := fmt.Sprintf("%s (%s)\ncontainer logs:", p.status, time.Since(*p.started).Truncate(time.Second))
	if p.done {
		return out
	}

	for l := range p.log.Iter() {
		out = out + "\n"
		if l.stderr {
			out = out + stderrStyle(l.line)
		} else {
			out = out + l.line
		}
	}

	return out
}

func (p *podman3RunTask) Run() (task.Result, error) {
	p.restartTimer("connecting to podman")

	ctx, err := bindings.NewConnection(context.Background(), getPodmanSocket())
	if err != nil {
		return task.Result{}, fmt.Errorf("error connecting to podman: %v", err)
	}

	p.restartTimer(fmt.Sprintf("pulling image %s to run %s %s", p.opts.Image, p.opts.Name.Before, p.opts.Name.Suffix))

	t := true
	_, err = images.Pull(ctx, p.opts.Image, &images.PullOptions{Quiet: &t})
	if err != nil {
		return task.Result{}, fmt.Errorf("error pulling image %s: %v", p.opts.Image, err)
	}

	p.restartTimer(fmt.Sprintf("booting container to %s %s", p.opts.Name.Before, p.opts.Name.Suffix))

	s := specgen.NewSpecGenerator(p.opts.Image, false)
	s.Entrypoint = p.opts.Entrypoint
	s.Command = p.opts.Command
	s.User = p.opts.User
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
		fmt.Println("⚠️  " + warning)
	}

	err = containers.Start(ctx, r.ID, &containers.StartOptions{})
	if err != nil {
		return task.Result{}, fmt.Errorf("error starting container: %s", err)
	}

	p.restartTimer(fmt.Sprintf("%s %s", p.opts.Name.During, p.opts.Name.Suffix))

	err = containers.Attach(ctx, r.ID, nil, p.log.Stdout(), p.log.Stderr(), nil, nil)
	if err != nil {
		return task.Result{}, fmt.Errorf("error attaching to container stdout: %v", err)
	}

	exitCode, err := containers.Wait(ctx, r.ID, &containers.WaitOptions{
		Condition: []define.ContainerStatus{define.ContainerStateStopped},
	})
	if err != nil {
		return task.Result{}, fmt.Errorf("error waiting for container to start: %v", err)
	}

	p.done = true

	if exitCode != 0 {
		return task.Result{}, fmt.Errorf("%s %s exited with code %d", p.opts.Name.Before, p.opts.Name.Suffix, exitCode)
	}

	err = containers.Remove(ctx, r.ID, nil)
	if err != nil {
		return task.Result{}, fmt.Errorf("error cleaning up container: %v", err)
	}

	p.status = fmt.Sprintf("%s %s", p.opts.Name.After, p.opts.Name.Suffix)

	return task.Result{Icon: task.ResultIconSuccess}, nil
}

func (p *podman3RunTask) restartTimer(status string) {
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

func getPodman3Mount(vol Volume) (specs.Mount, error) {
	source, err := filepath.Abs(vol.Source)
	if err != nil {
		return specs.Mount{}, err
	}

	return specs.Mount{
		Destination: vol.Destination,
		Source:      source,
		Options:     vol.Options,
		Type:        "bind",
	}, nil
}
