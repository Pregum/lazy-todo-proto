package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

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

	detailStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#25A065")).
			Padding(1)

	focusedBorder = lipgloss.Border{
		Top:         "━",
		Bottom:      "━",
		Left:        "┃",
		Right:       "┃",
		TopLeft:     "┏",
		TopRight:    "┓",
		BottomLeft:  "┗",
		BottomRight: "┛",
	}

	normalBorder = lipgloss.RoundedBorder()
)

type Task struct {
	title       string
	done        bool
	description string
	createdAt   time.Time
	updatedAt   time.Time
}

type FilterState int

const (
	ShowAll FilterState = iota
	ShowActive
	ShowCompleted
)

// 操作の種類を定義
type ActionType int

const (
	Add ActionType = iota
	Delete
	Edit
	Toggle
)

// 操作の履歴を保存する構造体
type Action struct {
	actionType ActionType
	task       Task
	index      int
}

type Model struct {
	panes      []string
	activePane int
	tasks      []Task
	cursor     int
	inputMode  bool
	input      string
	filePath   string
	editMode   bool
	editIndex  int
	filter     FilterState
	history    []Action
	redoStack  []Action
	width      int
	height     int
	focusPane  int // 0: タスクリスト, 1: 詳細ビュー
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
				title:     line,
				done:      done,
				createdAt: time.Now(),
				updatedAt: time.Now(),
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

func (m Model) filteredTasks() []Task {
	switch m.filter {
	case ShowActive:
		var active []Task
		for _, task := range m.tasks {
			if !task.done {
				active = append(active, task)
			}
		}
		return active
	case ShowCompleted:
		var completed []Task
		for _, task := range m.tasks {
			if task.done {
				completed = append(completed, task)
			}
		}
		return completed
	default:
		return m.tasks
	}
}

func (m *Model) addToHistory(action Action) {
	m.history = append(m.history, action)
	m.redoStack = nil // 新しい操作が行われたので、redoスタックをクリア
}

func (m *Model) adjustCursor() {
	if len(m.filteredTasks()) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(m.filteredTasks()) {
		m.cursor = len(m.filteredTasks()) - 1
	}
}

func (m *Model) undo() {
	if len(m.history) == 0 {
		return
	}

	lastAction := m.history[len(m.history)-1]
	m.history = m.history[:len(m.history)-1]
	m.redoStack = append(m.redoStack, lastAction)

	switch lastAction.actionType {
	case Add:
		m.tasks = append(m.tasks[:lastAction.index], m.tasks[lastAction.index+1:]...)
	case Delete:
		m.tasks = append(m.tasks[:lastAction.index], append([]Task{lastAction.task}, m.tasks[lastAction.index:]...)...)
	case Edit:
		m.tasks[lastAction.index] = lastAction.task
	case Toggle:
		m.tasks[lastAction.index].done = !m.tasks[lastAction.index].done
	}

	m.adjustCursor()
	if err := m.saveTasks(); err != nil {
		fmt.Printf("Error saving tasks: %v\n", err)
	}
}

func (m *Model) redo() {
	if len(m.redoStack) == 0 {
		return
	}

	nextAction := m.redoStack[len(m.redoStack)-1]
	m.redoStack = m.redoStack[:len(m.redoStack)-1]
	m.history = append(m.history, nextAction)

	switch nextAction.actionType {
	case Add:
		m.tasks = append(m.tasks[:nextAction.index], append([]Task{nextAction.task}, m.tasks[nextAction.index:]...)...)
	case Delete:
		m.tasks = append(m.tasks[:nextAction.index], m.tasks[nextAction.index+1:]...)
	case Edit:
		m.tasks[nextAction.index] = nextAction.task
	case Toggle:
		m.tasks[nextAction.index].done = !m.tasks[nextAction.index].done
	}

	m.adjustCursor()
	if err := m.saveTasks(); err != nil {
		fmt.Printf("Error saving tasks: %v\n", err)
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.inputMode {
			switch msg.String() {
			case "enter":
				if m.input != "" {
					if m.editMode {
						oldTask := m.tasks[m.editIndex]
						m.tasks[m.editIndex].title = m.input
						m.addToHistory(Action{
							actionType: Edit,
							task:       oldTask,
							index:      m.editIndex,
						})
						m.editMode = false
					} else {
						newTask := Task{title: m.input}
						m.tasks = append(m.tasks, newTask)
						m.addToHistory(Action{
							actionType: Add,
							task:       newTask,
							index:      len(m.tasks) - 1,
						})
					}
					m.input = ""
					if err := m.saveTasks(); err != nil {
						fmt.Printf("Error saving tasks: %v\n", err)
					}
				}
				m.inputMode = false
			case "esc":
				m.inputMode = false
				m.editMode = false
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
		case "tab", "h", "l":
			m.focusPane = (m.focusPane + 1) % 2
		case "n":
			m.inputMode = true
			m.editMode = false
		case "e":
			if len(m.tasks) > 0 {
				m.inputMode = true
				m.editMode = true
				m.editIndex = m.cursor
				m.input = m.tasks[m.cursor].title
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filteredTasks())-1 {
				m.cursor++
			}
		case " ":
			if len(m.tasks) > 0 {
				m.tasks[m.cursor].done = !m.tasks[m.cursor].done
				m.addToHistory(Action{
					actionType: Toggle,
					task:       m.tasks[m.cursor],
					index:      m.cursor,
				})
				if err := m.saveTasks(); err != nil {
					fmt.Printf("Error saving tasks: %v\n", err)
				}
			}
		case "d":
			if len(m.tasks) > 0 {
				deletedTask := m.tasks[m.cursor]
				m.addToHistory(Action{
					actionType: Delete,
					task:       deletedTask,
					index:      m.cursor,
				})
				m.tasks = append(m.tasks[:m.cursor], m.tasks[m.cursor+1:]...)
				m.adjustCursor()
				if err := m.saveTasks(); err != nil {
					fmt.Printf("Error saving tasks: %v\n", err)
				}
			}
		case "a":
			m.filter = ShowAll
			m.adjustCursor()
		case "c":
			m.filter = ShowCompleted
			m.adjustCursor()
		case "t":
			m.filter = ShowActive
			m.adjustCursor()
		case "u":
			m.undo()
		case "r":
			m.redo()
		}
	}
	return m, nil
}

func (m Model) View() string {
	// タイトルバー
	title := titleStyle.Render(" Lazy Todo ")

	// フィルター状態の表示
	filterText := "All"
	if m.filter == ShowActive {
		filterText = "Active"
	} else if m.filter == ShowCompleted {
		filterText = "Completed"
	}
	filterStatus := statusStyle.Render(fmt.Sprintf(" Filter: %s ", filterText))

	// ステータスバー
	status := statusStyle.Render(" Status: Ready | n: New Task | e: Edit | ↑/k: Up | ↓/j: Down | Space: Toggle | d: Delete | a: All | t: Active | c: Completed | u: Undo | r: Redo | Tab/h/l: Switch Pane | q: Quit ")

	// タスクリストの表示
	var taskList strings.Builder
	if m.inputMode {
		if m.editMode {
			taskList.WriteString(fmt.Sprintf("Editing task: %s_", m.input))
		} else {
			taskList.WriteString("> " + m.input + "_")
		}
	} else {
		filteredTasks := m.filteredTasks()
		for i, task := range filteredTasks {
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

	// 詳細ビューの表示
	var detail strings.Builder
	if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
		task := m.tasks[m.cursor]
		detail.WriteString(fmt.Sprintf("Title: %s\n", task.title))
		detail.WriteString(fmt.Sprintf("Status: %s\n", map[bool]string{true: "Done", false: "Active"}[task.done]))
		if task.description != "" {
			detail.WriteString(fmt.Sprintf("\nDescription:\n%s\n", task.description))
		}
		detail.WriteString(fmt.Sprintf("\nCreated: %s\n", task.createdAt.Format("2006-01-02 15:04:05")))
		detail.WriteString(fmt.Sprintf("Updated: %s\n", task.updatedAt.Format("2006-01-02 15:04:05")))
	} else {
		detail.WriteString("No task selected")
	}

	// ペインのスタイル設定
	listBorder := normalBorder
	detailBorder := normalBorder
	if m.focusPane == 0 {
		listBorder = focusedBorder
	} else {
		detailBorder = focusedBorder
	}

	listContent := lipgloss.NewStyle().
		Border(listBorder).
		BorderForeground(lipgloss.Color("#25A065")).
		Width(m.width/2 - 4).
		Render(taskList.String())

	detailContent := lipgloss.NewStyle().
		Border(detailBorder).
		BorderForeground(lipgloss.Color("#25A065")).
		Width(m.width/2 - 4).
		Render(detail.String())

	// 2ペインレイアウトの組み立て
	mainContent := lipgloss.JoinHorizontal(lipgloss.Left, listContent, detailContent)

	// 最終的なレイアウトの組み立て
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		title,
		filterStatus,
		mainContent,
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
		editMode:   false,
		editIndex:  0,
		filter:     ShowAll,
		history:    []Action{},
		redoStack:  []Action{},
		focusPane:  0,
	}

	// タスクの読み込み
	if err := m.loadTasks(); err != nil {
		fmt.Printf("Error loading tasks: %v\n", err)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
