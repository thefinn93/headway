package task

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type ResultIcon string

const (
	ResultIconSuccess   = ResultIcon("✅")
	ResultIconError     = ResultIcon("⚠️")
	ResultIconUnchanged = ResultIcon("☑️")
)

type Task interface {
	// View returns the status of the task
	View() string

	// Run is called from a goroutine when the task should execute
	Run() (Result, error)
}

type Result struct {
	Message string
	Icon    ResultIcon
}

// Execute a task and show it's status to the user
func Execute(task Task) {
	m := newModel(task)

	p := tea.NewProgram(&m)

	go func() {
		result, err := m.task.Run()
		if err != nil {
			m.err = err
		}
		m.result = result
		p.Quit()
	}()

	if err := p.Start(); err != nil {
		fmt.Println(ErrorStyle(fmt.Sprintf("%s  %s", ResultIconError, err)))
		os.Exit(1)
	}

	if m.err != nil {
		fmt.Println(ErrorStyle(fmt.Sprintf("%s  %s", ResultIconError, m.err)))
		os.Exit(1)
	}

	emoji := " "
	message := m.task.View()

	if m.result.Icon != "" {
		emoji = string(m.result.Icon)
	}

	if m.result.Message != "" {
		message = m.result.Message
	}

	fmt.Println(TextStyle(fmt.Sprintf("%s %s", emoji, message)))

	if m.quit {
		os.Exit(1)
	}
}
