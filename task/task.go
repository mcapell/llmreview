package task

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/mcapell/llmreview/llm"
	"github.com/mcapell/llmreview/llm/types"
	"github.com/mcapell/llmreview/question"
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

func (t *Task) Run(ctx context.Context) error {
	questions, err := question.ParseQuestionsFromPath(t.DataPath)
	if err != nil {
		return fmt.Errorf("error loading questions: %w", err)
	}

	for _, model := range t.Models {
		cli, err := llm.New(model)
		if err != nil {
			return err
		}

		for _, q := range questions {
			slog.Debug(fmt.Sprintf("processing question: %s using model %s", q.Path, cli))
			if t.resultExist(q) {
				slog.Debug("result already exist; ignoring question")
				continue
			}

			response, err := cli.Chat(ctx, types.Message{Prompt: t.Prompt, Text: q.Text})
			if err != nil {
				return fmt.Errorf("error from %s: %w", model, err)
			}

			if err := t.storeResult(q, response); err != nil {
				return fmt.Errorf("error storing result: %w", err)
			}
		}
	}

	return nil
}

func (t *Task) resultExist(q question.Question) bool {
	f, err := os.Open(t.ResultPath + q.Path)
	if err != nil {
		return false
	}
	defer f.Close()

	return true

}

func (t *Task) storeResult(q question.Question, result string) error {
	p := t.ResultPath + q.Path
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
