package task

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/key"
)

var (
	quitBindings = []key.Binding{
		key.NewBinding(key.WithKeys("q", "esc")),
		key.NewBinding(key.WithKeys("ctrl+c")),
	}
)

type model struct {
	task    Task
	spinner spinner.Model
	done    bool
	err     error
	result  Result
	quit bool
}

func newModel(task Task) model {
	s := spinner.New(spinner.WithSpinner(spinner.Dot))
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

	return model{
		spinner: s,
		task:    task,
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, quitBindings...) {
			m.quit = true
			return m, tea.Quit
		}
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)

		if m.done {
			cmd = tea.Batch(cmd, tea.Quit)
		}
		return m, cmd

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	return " " + m.spinner.View() + TextStyle(m.task.View())
}
