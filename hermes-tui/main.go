package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"hermes-tui/client"
	"hermes-tui/ui"
)

func main() {
	hermesClient := client.New("https://hermes-agent.mayuresh2k4.workers.dev", "default")

	model := ui.NewModel(hermesClient)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}