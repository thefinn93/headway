package task

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type Task interface {
	// View returns the status of the task
	View() string

	// Run is called from a goroutine when the task should execute
	Run() error
}

// Execute a task and show it's status to the user
func Execute(task Task) {
	m := newModel(task)

	p := tea.NewProgram(m)

	go func() {
		err := m.task.Run()
		if err != nil {
			m.err = err
		}
		p.Quit()
	}()

	if err := p.Start(); err != nil {
		fmt.Println(ErrorStyle(fmt.Sprintf("   %s", err)))
		os.Exit(1)
	}

	if m.err != nil {
		fmt.Println(ErrorStyle(fmt.Sprintf("  %s", m.err)))
		os.Exit(1)
	}

	fmt.Println(TextStyle(padding + m.task.View()))
}
