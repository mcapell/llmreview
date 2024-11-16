package task

import (
	"context"
	"fmt"
	"log/slog"

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
			slog.Debug(fmt.Sprintf("processing question: %s", q.Path))
			if question.ResultExists(t.ResultPath, q) {
				slog.Debug("result already exist; ignoring question")
				continue
			}

			response, err := cli.Chat(ctx, types.Message{Prompt: t.Prompt, Text: q.Text})
			if err != nil {
				return fmt.Errorf("error from %s: %w", model, err)
			}

			if err := question.StoreResult(t.ResultPath, q, response); err != nil {
				return fmt.Errorf("error storing result: %w", err)
			}
		}
	}

	return nil
}
