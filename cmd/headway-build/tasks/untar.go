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

func (u *UntarTask) Run() error {
	f, err := os.Open(u.source)
	if err != nil {
		return err
	}

	tr := tar.NewReader(f)
	i := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		i++

		filename := filepath.Join(u.dest, hdr.Name)

		u.status = fmt.Sprintf("untaring %s (%d - %s)", u.source, i, filename)

		err = os.Mkdir(filepath.Dir(filename), 0755)
		if err != nil && !os.IsExist(err) {
			return err
		}

		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err = io.Copy(file, tr); err != nil {
			return err
		}
		file.Close()
	}

	u.status = fmt.Sprintf("untaring %s (%d files)", u.source, i)

	return nil
}
