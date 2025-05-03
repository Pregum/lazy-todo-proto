package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/Pregum/lazy-todo-proto/internal/domain"
	"github.com/Pregum/lazy-todo-proto/internal/infrastructure"
	"github.com/Pregum/lazy-todo-proto/pkg/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	panes      []string
	activePane int
	tasks      []domain.Task
	cursor     int
	inputMode  bool
	input      string
	storage    *infrastructure.Storage
	editMode   bool
	editIndex  int
	filter     domain.FilterState
	history    []domain.Action
	redoStack  []domain.Action
	width      int
	height     int
	focusPane  int
	searchMode bool
	searchText string
}

func NewModel() Model {
	storage := infrastructure.NewStorage("todo.txt")
	tasks, err := storage.LoadTasks()
	if err != nil {
		fmt.Printf("Error loading tasks: %v\n", err)
	}

	return Model{
		panes:      []string{"Tasks", "Projects", "Settings"},
		activePane: 0,
		tasks:      tasks,
		cursor:     0,
		inputMode:  false,
		input:      "",
		storage:    storage,
		editMode:   false,
		editIndex:  0,
		filter:     domain.ShowAll,
		history:    []domain.Action{},
		redoStack:  []domain.Action{},
		focusPane:  0,
	}
}

func (m *Model) addToHistory(action domain.Action) {
	m.history = append(m.history, action)
	m.redoStack = nil
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

	switch lastAction.ActionType {
	case domain.Add:
		m.tasks = append(m.tasks[:lastAction.Index], m.tasks[lastAction.Index+1:]...)
	case domain.Delete:
		m.tasks = append(m.tasks[:lastAction.Index], append([]domain.Task{lastAction.Task}, m.tasks[lastAction.Index:]...)...)
	case domain.Edit:
		m.tasks[lastAction.Index] = lastAction.Task
	case domain.Toggle:
		m.tasks[lastAction.Index].Done = !m.tasks[lastAction.Index].Done
	}

	m.adjustCursor()
	if err := m.storage.SaveTasks(m.tasks); err != nil {
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

	switch nextAction.ActionType {
	case domain.Add:
		m.tasks = append(m.tasks[:nextAction.Index], append([]domain.Task{nextAction.Task}, m.tasks[nextAction.Index:]...)...)
	case domain.Delete:
		m.tasks = append(m.tasks[:nextAction.Index], m.tasks[nextAction.Index+1:]...)
	case domain.Edit:
		m.tasks[nextAction.Index] = nextAction.Task
	case domain.Toggle:
		m.tasks[nextAction.Index].Done = !m.tasks[nextAction.Index].Done
	}

	m.adjustCursor()
	if err := m.storage.SaveTasks(m.tasks); err != nil {
		fmt.Printf("Error saving tasks: %v\n", err)
	}
}

func (m Model) filteredTasks() []domain.Task {
	var filtered []domain.Task

	switch m.filter {
	case domain.ShowActive:
		for _, task := range m.tasks {
			if !task.Done {
				filtered = append(filtered, task)
			}
		}
	case domain.ShowCompleted:
		for _, task := range m.tasks {
			if task.Done {
				filtered = append(filtered, task)
			}
		}
	default:
		filtered = m.tasks
	}

	if m.searchText != "" {
		var searchFiltered []domain.Task
		for _, task := range filtered {
			if strings.Contains(strings.ToLower(task.Title), strings.ToLower(m.searchText)) {
				searchFiltered = append(searchFiltered, task)
			}
		}
		return searchFiltered
	}

	return filtered
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
		if m.searchMode {
			switch msg.String() {
			case "enter":
				m.searchMode = false
			case "esc":
				m.searchMode = false
				m.searchText = ""
			case "backspace":
				if len(m.searchText) > 0 {
					m.searchText = m.searchText[:len(m.searchText)-1]
				}
			case "delete":
				if len(m.searchText) > 0 {
					m.searchText = m.searchText[:len(m.searchText)-1]
				}
			case "ctrl+w":
				words := strings.Fields(m.searchText)
				if len(words) > 0 {
					words = words[:len(words)-1]
					m.searchText = strings.Join(words, " ")
				} else {
					m.searchText = ""
				}
			case "ctrl+u":
				m.searchText = ""
			default:
				if msg.Type == tea.KeyRunes {
					m.searchText += string(msg.Runes)
				}
			}
			return m, nil
		}

		if m.inputMode {
			switch msg.String() {
			case "enter":
				if m.input != "" {
					if m.editMode {
						oldTask := m.tasks[m.editIndex]
						m.tasks[m.editIndex].Title = m.input
						m.addToHistory(domain.Action{
							ActionType: domain.Edit,
							Task:       oldTask,
							Index:      m.editIndex,
						})
						m.editMode = false
					} else {
						newTask := domain.Task{
							Title:     m.input,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						}
						m.tasks = append(m.tasks, newTask)
						m.addToHistory(domain.Action{
							ActionType: domain.Add,
							Task:       newTask,
							Index:      len(m.tasks) - 1,
						})
					}
					m.input = ""
					if err := m.storage.SaveTasks(m.tasks); err != nil {
						fmt.Printf("Error saving tasks: %v\n", err)
					}
				}
				m.inputMode = false
			case "esc":
				m.inputMode = false
				m.editMode = false
				m.input = ""
			case "backspace":
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			case "delete":
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			case "ctrl+w":
				words := strings.Fields(m.input)
				if len(words) > 0 {
					words = words[:len(words)-1]
					m.input = strings.Join(words, " ")
				} else {
					m.input = ""
				}
			case "ctrl+u":
				m.input = ""
			default:
				if msg.Type == tea.KeyRunes {
					m.input += string(msg.Runes)
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			if err := m.storage.SaveTasks(m.tasks); err != nil {
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
				m.input = m.tasks[m.cursor].Title
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
				m.tasks[m.cursor].Done = !m.tasks[m.cursor].Done
				m.addToHistory(domain.Action{
					ActionType: domain.Toggle,
					Task:       m.tasks[m.cursor],
					Index:      m.cursor,
				})
				if err := m.storage.SaveTasks(m.tasks); err != nil {
					fmt.Printf("Error saving tasks: %v\n", err)
				}
			}
		case "d":
			if len(m.tasks) > 0 {
				deletedTask := m.tasks[m.cursor]
				m.addToHistory(domain.Action{
					ActionType: domain.Delete,
					Task:       deletedTask,
					Index:      m.cursor,
				})
				m.tasks = append(m.tasks[:m.cursor], m.tasks[m.cursor+1:]...)
				m.adjustCursor()
				if err := m.storage.SaveTasks(m.tasks); err != nil {
					fmt.Printf("Error saving tasks: %v\n", err)
				}
			}
		case "a":
			m.filter = domain.ShowAll
			m.adjustCursor()
		case "c":
			m.filter = domain.ShowCompleted
			m.adjustCursor()
		case "t":
			m.filter = domain.ShowActive
			m.adjustCursor()
		case "u":
			m.undo()
		case "r":
			m.redo()
		case "/":
			m.searchMode = true
			m.searchText = ""
		case "esc":
			if m.searchText != "" {
				m.searchText = ""
				m.adjustCursor()
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	title := ui.TitleStyle.Render(" Lazy Todo ")

	filterText := "All"
	if m.filter == domain.ShowActive {
		filterText = "Active"
	} else if m.filter == domain.ShowCompleted {
		filterText = "Completed"
	}
	filterStatus := ui.StatusStyle.Render(fmt.Sprintf(" Filter: %s ", filterText))

	if m.searchText != "" {
		filterStatus = ui.StatusStyle.Render(fmt.Sprintf(" Filter: %s | Search: %s ", filterText, m.searchText))
	}

	if m.searchMode {
		filterStatus = ui.StatusStyle.Render(fmt.Sprintf(" Filter: %s | Search: %s_", filterText, m.searchText))
	}

	status := ui.StatusStyle.Render(" Status: Ready | n: New Task | e: Edit | ↑/k: Up | ↓/j: Down | Space: Toggle | d: Delete | a: All | t: Active | c: Completed | u: Undo | r: Redo | /: Search | Esc: Clear Search | Tab/h/l: Switch Pane | q: Quit ")

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
			if task.Done {
				done = "✓"
			}
			taskList.WriteString(fmt.Sprintf("%s [%s] %s\n", cursor, done, task.Title))
		}
	}

	var detail strings.Builder
	if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
		task := m.tasks[m.cursor]
		detail.WriteString(fmt.Sprintf("Title: %s\n", task.Title))
		detail.WriteString(fmt.Sprintf("Status: %s\n", map[bool]string{true: "Done", false: "Active"}[task.Done]))
		if task.Description != "" {
			detail.WriteString(fmt.Sprintf("\nDescription:\n%s\n", task.Description))
		}
		detail.WriteString(fmt.Sprintf("\nCreated: %s\n", task.CreatedAt.Format("2006-01-02 15:04:05")))
		detail.WriteString(fmt.Sprintf("Updated: %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05")))
	} else {
		detail.WriteString("No task selected")
	}

	listBorder := ui.NormalBorder
	detailBorder := ui.NormalBorder
	if m.focusPane == 0 {
		listBorder = ui.FocusedBorder
	} else {
		detailBorder = ui.FocusedBorder
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

	mainContent := lipgloss.JoinHorizontal(lipgloss.Left, listContent, detailContent)

	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		title,
		filterStatus,
		mainContent,
		lipgloss.NewStyle().Height(1).Render(""),
		status,
	)
} 