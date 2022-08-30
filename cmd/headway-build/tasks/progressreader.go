package tasks

import (
	"io"
)

// ProgressReader allows you to monitor the progress of reading through an io.Reader
type ProgressReader struct {
	total    int
	progress int
	reader   io.Reader
}

// NewProgressReader creates a ProgressReader
func NewProgressReader(total int, reader io.Reader) *ProgressReader {
	return &ProgressReader{
		total:  total,
		reader: reader,
	}
}

// GetPercent returns the current progress as a float
func (pw *ProgressReader) GetPercent() float64 {
	return float64(pw.progress) / float64(pw.total)
}

// Done returns true if the progress is at least 100%
func (pw *ProgressReader) Done() bool {
	return pw.progress >= pw.total
}

func (pw *ProgressReader) Read(p []byte) (int, error) {
	s, err := pw.reader.Read(p)
	pw.progress += s
	return s, err
}
