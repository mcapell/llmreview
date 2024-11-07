package question

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	Divider = "---"

	SectionNone     = "none"
	SectionMetadata = "metadata"
	SectionQuestion = "question"
	SectionAnswer   = "answer"
)

func ParseQuestionsFromPath(path string) ([]Question, error) {
	var questions []Question
	if err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		question, err := parseQuestionFromFile(path)
		if err != nil {
			return fmt.Errorf("error parsing question: %w", err)
		}

		questions = append(questions, *question)

		return nil
	}); err != nil {
		return nil, err
	}

	return questions, nil
}

func parseQuestionFromFile(path string) (*Question, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return parseQuestion(bytes.NewReader(content))
}

func parseQuestion(content io.Reader) (*Question, error) {
	scanner := bufio.NewScanner(content)
	section := SectionNone

	question := strings.Builder{}
	answer := strings.Builder{}
	metadata := map[string]string{}
	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == Divider {
			section = nextSection(section)
			continue
		}

		switch section {
		case SectionMetadata:
			parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
			if len(parts) > 1 {
				metadata[parts[0]] = parts[1]
			}
		case SectionQuestion:
			question.WriteString(line + " ")
		case SectionAnswer:
			answer.WriteString(line)
		}

	}
	return &Question{
		Text:     question.String(),
		Solution: answer.String(),
		Metadata: metadata,
	}, nil
}

func nextSection(current string) string {
	switch current {
	case SectionNone:
		return SectionMetadata
	case SectionMetadata:
		return SectionQuestion
	default:
		return SectionAnswer
	}
}
