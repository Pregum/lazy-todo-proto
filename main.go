package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// スタイル定義
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#3C3C3C")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#1A1A1A")).
			Padding(0, 1)
)

type Model struct {
	panes      []string
	activePane int
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.activePane = (m.activePane + 1) % len(m.panes)
		}
	}
	return m, nil
}

func (m Model) View() string {
	// タイトルバー
	title := titleStyle.Render(" Lazy Todo ")

	// ステータスバー
	status := statusStyle.Render(" Status: Ready ")

	// ペインのヘッダー
	var paneHeaders string
	for i, pane := range m.panes {
		if i == m.activePane {
			paneHeaders += selectedStyle.Render(" " + pane + " ")
		} else {
			paneHeaders += normalStyle.Render(" " + pane + " ")
		}
	}

	// メインコンテンツエリア
	content := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#25A065")).
		Padding(1).
		Render("Content Area")

	// レイアウトの組み立て
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		title,
		paneHeaders,
		content,
		lipgloss.NewStyle().Height(1).Render(""),
		status,
	)
}

func main() {
	m := Model{
		panes:      []string{"Tasks", "Projects", "Settings"},
		activePane: 0,
	}
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
