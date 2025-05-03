package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct{}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00")).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder())

	return style.Render("Hello, World!") + "\n\n" +
		lipgloss.NewStyle().Faint(true).Render("Press 'q' to quit")
}

func main() {
	p := tea.NewProgram(Model{})
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
