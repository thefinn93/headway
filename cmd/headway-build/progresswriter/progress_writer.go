package progresswriter

import (
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type ErrMsg struct {
	Err error
}

type ProgressWriter struct {
	total      int
	downloaded int
	file       *os.File
	reader     io.Reader
	p          *tea.Program
}

// New creates a ProgressWriter
func New(total int, file *os.File, reader io.Reader, p *tea.Program) ProgressWriter {
	return ProgressWriter{
		total:      total,
		file:       file,
		reader:     reader,
		p:          p,
	}
}

func (pw *ProgressWriter) Start() {
	// TeeReader calls pw.Write() each time a new response is received
	_, err := io.Copy(pw.file, io.TeeReader(pw.reader, pw))
	if err != nil {
		if pw.p != nil {
			pw.p.Send(ErrMsg{err})
		}
	}
}

func (pw *ProgressWriter) GetPercent() float64 {
	return float64(pw.downloaded) / float64(pw.total)
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	pw.downloaded += len(p)
	return len(p), nil
}
