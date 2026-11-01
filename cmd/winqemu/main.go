package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"winqemu/internal/tui"
)

func main() {
	app, err := tui.NewAppModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing WinQEMU: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
