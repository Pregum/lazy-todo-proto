package main

import (
	"fmt"
	"strings"

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

	listStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#25A065")).
			Padding(1)
)

type Task struct {
	title string
	done  bool
}

type Model struct {
	panes      []string
	activePane int
	tasks      []Task
	cursor     int
	inputMode  bool
	input      string
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.inputMode {
			switch msg.String() {
			case "enter":
				if m.input != "" {
					m.tasks = append(m.tasks, Task{title: m.input})
					m.input = ""
				}
				m.inputMode = false
			case "esc":
				m.inputMode = false
				m.input = ""
			default:
				m.input += msg.String()
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.activePane = (m.activePane + 1) % len(m.panes)
		case "n":
			m.inputMode = true
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}
		case " ":
			if len(m.tasks) > 0 {
				m.tasks[m.cursor].done = !m.tasks[m.cursor].done
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	// タイトルバー
	title := titleStyle.Render(" Lazy Todo ")

	// ステータスバー
	status := statusStyle.Render(" Status: Ready | n: New Task | ↑/k: Up | ↓/j: Down | Space: Toggle | q: Quit ")

	// ペインのヘッダー
	var paneHeaders string
	for i, pane := range m.panes {
		if i == m.activePane {
			paneHeaders += selectedStyle.Render(" " + pane + " ")
		} else {
			paneHeaders += normalStyle.Render(" " + pane + " ")
		}
	}

	// タスクリストの表示
	var taskList strings.Builder
	if m.inputMode {
		taskList.WriteString("> " + m.input + "_")
	} else {
		for i, task := range m.tasks {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			done := " "
			if task.done {
				done = "✓"
			}
			taskList.WriteString(fmt.Sprintf("%s [%s] %s\n", cursor, done, task.title))
		}
	}

	// メインコンテンツエリア
	content := listStyle.Render(taskList.String())

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
		tasks:      []Task{},
		cursor:     0,
		inputMode:  false,
		input:      "",
	}
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
