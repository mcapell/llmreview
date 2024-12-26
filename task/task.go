package task

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Task struct {
	Name       string   `yaml:"name"`
	Models     []string `yaml:"models"`
	Prompt     string   `yaml:"prompt"`
	DataPath   string   `yaml:"data_path"`
	ResultPath string   `yaml:"result_path"`
}

type Tasks struct {
	Tasks []Task `yaml:"tasks"`
}

func (t *Task) ResultExist(path string) bool {
	return t.LoadResult(path) != ""
}

func (t *Task) LoadResult(path string) string {
	f, err := os.Open(t.ResultPath + path)
	if err != nil {
		return ""
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return ""
	}

	return string(b)
}

func (t *Task) StoreResult(path, result string) error {
	p := t.ResultPath + path
	if err := os.MkdirAll(filepath.Dir(p), os.ModePerm); err != nil {
		return fmt.Errorf("error creating the results path: %w", err)
	}

	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("error creating result file: %w", err)
	}
	defer f.Close()

	if _, err := io.WriteString(f, result); err != nil {
		return fmt.Errorf("error writing result: %w", err)
	}

	return nil
}
