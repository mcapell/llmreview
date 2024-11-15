package task

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

type Runner struct {
	config string
}

func NewRunner(configPath string) *Runner {
	return &Runner{
		config: configPath,
	}
}

func (r *Runner) Run() error {
	tasks, err := r.readTasks()
	if err != nil {
		return fmt.Errorf("error reading configuration file: %w", err)
	}

	return r.runTasks(tasks)
}

func (r *Runner) readTasks() ([]Task, error) {
	f, err := os.Open(r.config)
	if err != nil {
		return nil, err
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var tasks Tasks
	if err := yaml.Unmarshal(content, &tasks); err != nil {
		return nil, err
	}

	return tasks.Tasks, nil
}

func (r *Runner) runTasks(tasks []Task) error {
	ctx := context.Background()
	for _, task := range tasks {
		slog.Debug("running task", "task", task)
		if err := task.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}
