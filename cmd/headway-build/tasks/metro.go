package tasks

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/key"

	"github.com/headwaymaps/headway/cmd/headway-build/tasks/task"
)

var (
	listStyle = lipgloss.NewStyle().Margin(1, 2)
)

// Metro describes a particular metro region that Headway can be built for
type Metro struct {
	Name           string     // Name is the name of the metro region
	NormalizedName string     // NormalizedName is the normalized version of the name for quick typeahead searching
	Country        string     // Country is the two character country code of the country that the metro region resides in
	Coords         [4]float64 // Coords is two long/lat pairs that describe a box covering the metro area
	Population     int64      // Population is the population of the metro area, for ordering conflicts
}

type metroSelect struct {
	list      list.Model
	quit bool
	done bool
}


func (m Metro) Title() string       { return m.Name }
func (m Metro) Description() string { return m.Country }
func (m Metro) FilterValue() string { return m.Name }


func (t metroSelect) Init() tea.Cmd {
	return nil
}

func (t *metroSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, t.list.KeyMap.Quit, t.list.KeyMap.ForceQuit) {
				t.quit = true
				return t, tea.Quit
		}

		if msg.String() == "enter" {
			return t, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := listStyle.GetFrameSize()
		t.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	t.list, cmd = t.list.Update(msg)
	return t, cmd
}

func (t metroSelect) View() string {
	if t.done {
		if t.quit {
			return "no metro selected"
		}

		selected := t.list.SelectedItem().(Metro)
		return "selected " + selected.Name
	}
	return t.list.View()
}

func GetMetro(items []list.Item) Metro {
	m := metroSelect{
		list: list.New(items, list.NewDefaultDelegate(), 0, 0),
	}

	m.list.Title = "select metro area"
	m.list.SetShowFilter(true)

	if err := tea.NewProgram(&m, tea.WithAltScreen()).Start(); err != nil {
		fmt.Println(task.ErrorStyle(fmt.Sprintf("%s  %s", task.ResultIconError, err)))
		os.Exit(1)
	}

	if m.quit {
		os.Exit(1)
	}

	return m.list.SelectedItem().(Metro)
}
