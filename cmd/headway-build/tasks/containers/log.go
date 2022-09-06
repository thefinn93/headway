package containers

import (
	"bytes"
	"io"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

const logLineCount = 10

var stderrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#dd0000")).Render

type logger struct {
	sync.RWMutex
	items []logLine
}

type logLine struct {
	stderr bool
	line   string
}

type logWriter struct {
	logger *logger
	buf    []byte
	stderr bool
}

func (cl *logger) Iter() <-chan logLine {
	ch := make(chan logLine)

	go func() {
		cl.Lock()
		defer cl.Unlock()

		for _, value := range cl.items {
			ch <- value
		}

		close(ch)
	}()

	return ch
}

func (c *logWriter) Write(p []byte) (n int, err error) {
	lines := bytes.Split(p, []byte("\n"))
	lineCount := len(lines)
	for i, line := range lines {
		if i < lineCount-1 {
			if i == 0 {
				line = append(c.buf, line...)
			}

			c.logger.append(logLine{stderr: c.stderr, line: string(line)})
		} else {
			c.buf = line
		}
	}
	return len(p), nil
}

func (cl *logger) append(item logLine) {
	cl.Lock()
	defer cl.Unlock()

	cl.items = append(cl.items, item)

	if len(cl.items) >= logLineCount {
		offset := len(cl.items) - logLineCount
		cl.items = cl.items[offset:]
	}
}

// Stdout returns a writer for standard out logs
func (cl *logger) Stdout() io.Writer {
	return &logWriter{logger: cl}
}

// Stderr returns a writer for standard error logs
func (cl *logger) Stderr() io.Writer {
	return &logWriter{logger: cl, stderr: true}
}
