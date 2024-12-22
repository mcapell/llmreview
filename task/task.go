package task

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/mcapell/llmreview/llm"
	"github.com/mcapell/llmreview/llm/types"
	"github.com/mcapell/llmreview/question"
)

//go:embed gradingprompt.txt
var gradingPrompt string

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

type GradeResult struct {
	Grades []struct {
		Grade    int    `json:"grade,omitempty"`
		Category string `json:"category,omitempty"`
		Notes    string `json:"notes,omitempty"`
	} `json:"grades,omitempty"`
}

func (t *Task) Run(ctx context.Context) error {
	questions, err := question.ParseQuestionsFromPath(t.DataPath)
	if err != nil {
		return fmt.Errorf("error loading questions: %w", err)
	}

	// Use openAI as the correction model
	correctionCli, err := llm.New("openai")
	if err != nil {
		return fmt.Errorf("error creating correction model: %w", err)
	}

	for _, model := range t.Models {
		cli, err := llm.New(model)
		if err != nil {
			return err
		}

		for _, q := range questions {
			slog.Debug(fmt.Sprintf("processing question: %s using model %s", q.Path, cli))

			result, err := t.getResult(ctx, cli, q)
			if err != nil {
				return fmt.Errorf("error getting LLM result: %w", err)
			}

			slog.Debug(fmt.Sprintf("evaluating result: %s using model %s", q.Path, correctionCli))
			if err = t.gradeResult(ctx, correctionCli, q, result); err != nil {
				return fmt.Errorf("error evaluating LLM result: %w", err)
			}
		}
	}

	return nil
}

func (t *Task) getResult(ctx context.Context, cli llm.Client, q question.Question) (string, error) {
	if t.resultExist(q.QuestionPath()) {
		slog.Debug("result already exist; loading it")

		return t.loadResult(q.QuestionPath()), nil
	}

	response, err := cli.Chat(ctx, t.Prompt, []types.Message{{Content: q.Content}})
	if err != nil {
		return "", fmt.Errorf("error from %s: %w", cli, err)
	}

	if err := t.storeResult(q.QuestionPath(), response); err != nil {
		return "", fmt.Errorf("error storing result: %w", err)
	}

	return response, nil
}

func (t *Task) gradeResult(ctx context.Context, cli llm.Client, q question.Question, result string) error {
	if q.CorrectionType == "" {
		slog.Debug("correction does not exist; ignore evaluation")
		return nil
	}

	if t.resultExist(q.GradingPath()) {
		slog.Debug("evaluation already exist")
		return nil
	}

	response, err := cli.Chat(ctx, gradingPrompt+q.Correction, []types.Message{{Content: []types.Content{{Text: result}}}})
	if err != nil {
		return fmt.Errorf("error from %s: %w", cli, err)
	}

	var grade GradeResult
	if err := json.NewDecoder(strings.NewReader(response)).Decode(&grade); err != nil {
		return fmt.Errorf("error parsing grade result: %w", err)
	}

	if err := t.storeResult(q.GradingPath(), response); err != nil {
		return fmt.Errorf("error storing result: %w", err)
	}

	return nil
}

func (t *Task) resultExist(path string) bool {
	return t.loadResult(path) != ""
}

func (t *Task) loadResult(path string) string {
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

func (t *Task) storeResult(path, result string) error {
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
