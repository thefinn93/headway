package tasks

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/headwaymaps/headway/cmd/headway-build/tasks/task"
)

type UntarTask struct {
	source string
	dest   string
	status string
}

func Untar(source, dest string) {
	task.Execute(&UntarTask{
		source: source,
		dest:   dest,
		status: fmt.Sprintf("untaring %s", source),
	})
}

func (u UntarTask) View() string {
	return u.status
}

func (u *UntarTask) Run() (task.Result, error) {
	f, err := os.Open(u.source)
	if err != nil {
		return task.Result{}, fmt.Errorf("error reading %s: %v", u.source, err)
	}

	tr := tar.NewReader(f)
	i := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return task.Result{}, fmt.Errorf("error reading tar file: %v", err)
		}
		i++

		filename := filepath.Join(u.dest, hdr.Name)

		u.status = fmt.Sprintf("untaring %s (%d - %s)", u.source, i, filename)

		dir := filepath.Dir(filename)
		err = os.Mkdir(dir, 0755)
		if err != nil && !os.IsExist(err) {
			return task.Result{}, fmt.Errorf("error making directory %s: %v", dir, err)
		}

		file, err := os.Create(filename)
		if err != nil {
			return task.Result{}, fmt.Errorf("error writing to %s: %v", filename, err)
		}
		defer file.Close()

		if _, err = io.Copy(file, tr); err != nil {
			return task.Result{}, fmt.Errorf("error extracting file %s: %v", filename, err)
		}
		file.Close()
	}

	return task.Result{
		Icon: task.ResultIconSuccess,
		Message: fmt.Sprintf("untaring %s (%d files)", u.source, i),
	}, nil
}
