package tasks

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/headwaymaps/headway/cmd/headway-build/tasks/task"
)

var (
	errNoLastModifiedHeader = errors.New("no headers indicated server last modified time")
)

// DownloadTask downloads a file to the local disk
type DownloadTask struct {
	url      string
	dest     string
	progress *ProgressReader
	message  string
	changed bool
}

func Download(url, dest string) bool {
	t  := DownloadTask{
		url:  url,
		dest: dest,
	}
	task.Execute(&t)
	return t.changed
}

func (d DownloadTask) View() string {
	if d.message != "" {
		return d.message
	}

	if d.progress == nil {
		return fmt.Sprintf("downloading %s to %s", d.url, d.dest)
	}

	if d.progress.Done() {
		return fmt.Sprintf("downloaded %s to %s", d.url, d.dest)
	}

	return fmt.Sprintf("downloading %s to %s (%.2f%%)", d.url, d.dest, d.progress.GetPercent()*100)

}

func (d *DownloadTask) Run() error {
	existing, err := os.Stat(d.dest)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error checking target file %s: %v", d.dest, err)
	}

	resp, err := http.Head(d.url)
	if err != nil {
		return fmt.Errorf("error making http HEAD request to %s before downloading: %v", d.url, err)
	}

	if !shouldRedownload(resp.Header, existing) {
		d.message = fmt.Sprintf("not redownloading %s to %s", d.url, d.dest)
		return nil
	}

	resp, err = getResponse(d.url)
	if err != nil {
		return fmt.Errorf("error downloading %s: %v", d.url, err)
	}
	defer resp.Body.Close()

	file, err := os.Create(d.dest)
	if err != nil {
		return err
	}
	defer file.Close()

	d.progress = NewProgressReader(int(resp.ContentLength), resp.Body)

	if resp.ContentLength == 0 {
		return nil
	}

	d.changed = true

	f, err := os.Create(d.dest)
	if err != nil {
		return fmt.Errorf("error opening %s for writing: %v", d.dest, err)
	}
	defer f.Close()

	_, err = io.Copy(f, d.progress)
	if err != nil {
		return fmt.Errorf("error downloading %s to %s: %v", d.url, d.dest, err)
	}

	return nil
}

func getResponse(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("receiving status of %d for url: %s", resp.StatusCode, url)
	}
	return resp, nil
}

func shouldRedownload(headers http.Header, existing os.FileInfo) bool {
	lastModified, err := getLastModified(headers)
	if err != nil && err != errNoLastModifiedHeader {
		fmt.Println("error parsing Last-Modified value from server, not re-downloading: ", err)
		return false
	} else if err == nil {
		if existing == nil {
			return true
		}

		if existing.ModTime().After(lastModified) {
			return false
		}
	}

	// TODO: even if Last-Modified header is older, compare size on disk with Content-Length header to detect partial downloads

	return true
}

func getLastModified(headers http.Header) (time.Time, error) {
	if header := headers.Get("Last-Modified"); header != "" {
		return time.Parse("Mon, 2 Jan 2006 15:04:05 MST", header)
	}

	if header := headers.Get("X-Bz-Upload-Timestamp"); header != "" {
		mills, err := strconv.ParseInt(header, 10, 64)
		if err != nil {
			return time.Time{}, err
		}

		return time.UnixMicro(mills), nil
	}

	return time.Time{}, errNoLastModifiedHeader
}
