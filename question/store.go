package question

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func ResultExists(basepath string, question Question) bool {
	p := basepath + question.Path

	f, err := os.Open(p)
	if err != nil {
		return false
	}
	defer f.Close()

	return true
}

func StoreResult(basepath string, question Question, result string) error {
	p := basepath + question.Path
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
