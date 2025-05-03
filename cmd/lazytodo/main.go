package main

import (
	"fmt"
	"os"

	"github.com/Pregum/lazy-todo-proto/internal/app"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := app.NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
} 