package infrastructure

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Pregum/lazy-todo-proto/internal/domain"
)

type Storage struct {
	filePath string
}

func NewStorage(filePath string) *Storage {
	return &Storage{
		filePath: filePath,
	}
}

func (s *Storage) LoadTasks() ([]domain.Task, error) {
	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []domain.Task{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var tasks []domain.Task
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 {
			done := false
			if strings.HasPrefix(line, "x ") {
				done = true
				line = line[2:]
			}
			tasks = append(tasks, domain.Task{
				Title:     line,
				Done:      done,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}
	}
	return tasks, scanner.Err()
}

func (s *Storage) SaveTasks(tasks []domain.Task) error {
	file, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, task := range tasks {
		prefix := "x "
		if !task.Done {
			prefix = ""
		}
		_, err := fmt.Fprintf(file, "%s%s\n", prefix, task.Title)
		if err != nil {
			return err
		}
	}
	return nil
} 