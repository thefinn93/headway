package download

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/headwaymaps/headway/cmd/headway-build/progresswriter"
)

const (
	padding  = 2
	maxWidth = 80
)

var (
	p         *tea.Program
	textStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#dddddd")).Render
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

	errNoLastModifiedHeader = errors.New("no headers indicated server last modified time")
)

func getResponse(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("receiving status of %d for url: %s", resp.StatusCode, url)
	}
	return resp, nil
}

func finalPause() tea.Cmd {
	return tea.Tick(time.Millisecond*750, func(_ time.Time) tea.Msg {
		return nil
	})
}

type model struct {
	url     string
	dest    string
	pw      *progresswriter.ProgressWriter
	spinner spinner.Model
	err     error
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)

		if m.pw.GetPercent() >= 1 {
			cmd = tea.Batch(cmd, tea.Quit)
		}
		return m, cmd

	case progresswriter.ErrMsg:
		m.err = msg.Err
		return m, tea.Quit

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	if m.err != nil {
		return "Error downloading: " + m.err.Error() + "\n"
	}

	return m.spinner.View() + textStyle(fmt.Sprintf("downloading %s to %s (%.2f%%)", m.url, m.dest, m.pw.GetPercent()*100))
}

// Download a url to a unique filename. Will attempt to avoid downloading by checking for last-modified and similar headers
// returns true if the file was updated
func Download(url, dest string) bool {
	existing, err := os.Stat(dest)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	} else {
		resp, err := http.Head(url)
		if err != nil {
			log.Fatal("error making request: HEAD ", url, ": ", err)
		}

		if !shouldRedownload(resp.Header, existing) {
			fmt.Println("  not redownloading", url, "to", dest)
			return false
		}

		log.Println(dest, "exists but the server has a newer version, redownloading")
	}

	resp, err := getResponse(url)
	if err != nil {
		fmt.Println("could not get response", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	file, err := os.Create(dest)
	if err != nil {
		fmt.Println("could not create file: ", err)
		os.Exit(1)
	}
	defer file.Close()

	pw := progresswriter.New(int(resp.ContentLength), file, resp.Body, p)

	m := model{
		pw:      &pw,
		spinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
		url:     url,
		dest:    dest,
	}

	m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

	// Start the download
	go pw.Start()

	// Don't add TUI if the header doesn't include content size
	// it's impossible see progress without total
	if resp.ContentLength > 0 {
		// Start Bubble Tea
		p = tea.NewProgram(m)
		if err := p.Start(); err != nil {
			fmt.Println("error downloading: ", err)
			os.Exit(1)
		}
		fmt.Printf("  downloaded %s to %s", url, dest)
		fmt.Println()
	}

	return true
}

func shouldRedownload(headers http.Header, existing os.FileInfo) bool {
	lastModified, err := getLastModified(headers)
	if err != nil && err != errNoLastModifiedHeader {
		log.Println("error parsing Last-Modified value from server, not re-downloading: ", err)
		return false
	} else if err == nil {
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
