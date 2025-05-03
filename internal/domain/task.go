package domain

import "time"

type Task struct {
	Title       string
	Done        bool
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type FilterState int

const (
	ShowAll FilterState = iota
	ShowActive
	ShowCompleted
)

type ActionType int

const (
	Add ActionType = iota
	Delete
	Edit
	Toggle
)

type Action struct {
	ActionType ActionType
	Task       Task
	Index      int
} 