package question

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	correctionSuffix = "_correction"
	gradeSuffix      = "_grade"
)

var (
	FileNotFoundErr = errors.New("file not found")
)

func ParseQuestionsFromPath(path string) ([]Question, error) {
	var questions []Question
	filesParsed := map[string]bool{}

	if err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Check if the file was already seen
		if _, seen := filesParsed[path]; seen {
			return nil
		}

		// Only process question files, and create the correction file based on the original name
		var questionPath, corrPath string
		if !strings.Contains(path, correctionSuffix) {
			questionPath = path
			corrPath = addPathSuffix(path, correctionSuffix)
			filesParsed[questionPath] = true
			filesParsed[corrPath] = true
		}

		fmt.Printf("question path: %s - correction path: %s\n", questionPath, corrPath)

		question := Question{
			Path:         path,
			QuestionType: InputText,
		}

		questionInput, err := parseText(questionPath)
		if err != nil {
			return fmt.Errorf("error parsing question: %w", err)
		}

		question.Text = questionInput

		// Correction is optional. Add only if the file exist
		correctionInput, err := parseText(corrPath)
		if err != nil && !errors.Is(err, FileNotFoundErr) {
			return fmt.Errorf("error parsing question: %w", err)
		} else if err == nil {
			question.Correction = correctionInput
			question.CorrectionType = InputText
		}

		questions = append(questions, question)
		return nil
	}); err != nil {
		return nil, err
	}

	return questions, nil
}

func parseText(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", FileNotFoundErr
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return string(content), nil
}
