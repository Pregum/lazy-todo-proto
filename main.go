package main

import (
	"bufio"
	"fmt"
	"os"
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
	filePath   string
}

func (m *Model) loadTasks() error {
	file, err := os.Open(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // ファイルが存在しない場合は空のタスクリストで開始
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 {
			done := false
			if strings.HasPrefix(line, "x ") {
				done = true
				line = line[2:]
			}
			m.tasks = append(m.tasks, Task{
				title: line,
				done:  done,
			})
		}
	}
	return scanner.Err()
}

func (m *Model) saveTasks() error {
	file, err := os.Create(m.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, task := range m.tasks {
		prefix := "x "
		if !task.done {
			prefix = ""
		}
		_, err := fmt.Fprintf(file, "%s%s\n", prefix, task.title)
		if err != nil {
			return err
		}
	}
	return nil
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
					if err := m.saveTasks(); err != nil {
						fmt.Printf("Error saving tasks: %v\n", err)
					}
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
			if err := m.saveTasks(); err != nil {
				fmt.Printf("Error saving tasks: %v\n", err)
			}
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
				if err := m.saveTasks(); err != nil {
					fmt.Printf("Error saving tasks: %v\n", err)
				}
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
		filePath:   "todo.txt",
	}

	// タスクの読み込み
	if err := m.loadTasks(); err != nil {
		fmt.Printf("Error loading tasks: %v\n", err)
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
