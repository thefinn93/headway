package containers

import (
	"bytes"
	"io"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

const containerLogLineCount = 10

var stderrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#dd0000")).Render

type containerLogger struct {
	sync.RWMutex
	items []containerLogLine
}

type containerLogLine struct {
	stderr bool
	line   string
}

type containerLogWriter struct {
	logger *containerLogger
	buf    []byte
	stderr bool
}

func (cl *containerLogger) append(item containerLogLine) {
	cl.Lock()
	defer cl.Unlock()

	cl.items = append(cl.items, item)

	if len(cl.items) >= containerLogLineCount {
		offset := len(cl.items) - containerLogLineCount
		cl.items = cl.items[offset:]
	}
}

func (cl *containerLogger) Iter() <-chan containerLogLine {
	ch := make(chan containerLogLine)

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

func (c *containerLogWriter) Write(p []byte) (n int, err error) {
	lines := bytes.Split(p, []byte("\n"))
	lineCount := len(lines)
	for i, line := range lines {
		if i < lineCount-2 {
			if i == 0 {
				line = append(c.buf, line...)
			}

			c.logger.append(containerLogLine{stderr: c.stderr, line: string(line)})
		} else {
			c.buf = line
		}
	}
	return len(p), nil
}

// Stdout returns a writer for standard out logs
func (cl *containerLogger) Stdout() io.Writer {
	return &containerLogWriter{logger: cl}
}

// Stderr returns a writer for standard error logs
func (cl *containerLogger) Stderr() io.Writer {
	return &containerLogWriter{logger: cl, stderr: true}
}
