package task

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/mcapell/llmreview/llm"
	"github.com/mcapell/llmreview/llm/types"
	"github.com/mcapell/llmreview/question"
	"gopkg.in/yaml.v3"
)

type Runner struct {
	config string
	grader Grader
}

func NewRunner(configPath string) *Runner {
	return &Runner{
		config: configPath,
		grader: NewGrader(),
	}
}

func (r *Runner) Run() error {
	tasks, err := r.readTasks()
	if err != nil {
		return fmt.Errorf("error reading configuration file: %w", err)
	}

	ctx := context.Background()
	for _, task := range tasks {
		if err := r.runTask(ctx, task); err != nil {
			return err
		}
	}
	return nil
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

func (r *Runner) runTask(ctx context.Context, t Task) error {
	slog.Debug("running task", "task", t)

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

			// Get question response
			result := t.LoadResult(q.ResultPath())
			if result == "" {
				result, err = r.getQuestionResult(ctx, cli, t.Prompt, q)
				if err != nil {
					return fmt.Errorf("error getting LLM result: %w", err)
				}

				if err := t.StoreResult(q.ResultPath(), result); err != nil {
					return fmt.Errorf("error storing result: %w", err)
				}
			}

			if q.CorrectionType == "" {
				slog.Debug("correction does not exist; ignore evaluation")
				continue
			}

			// Grade response
			slog.Debug(fmt.Sprintf("evaluating result: %s", q.Path))
			if !t.ResultExist(q.GradingPath()) {
				gradeResult, err := r.grader.GradeQuestionResult(ctx, q, result)
				if err != nil {
					return fmt.Errorf("error evaluating LLM result: %w", err)
				}

				if err := t.StoreResult(q.GradingPath(), gradeResult); err != nil {
					return fmt.Errorf("error storing grading: %w", err)
				}
			}
		}
	}

	return nil
}

func (r *Runner) getQuestionResult(ctx context.Context, cli llm.Client, prompt string, q question.Question) (string, error) {
	response, err := cli.Chat(ctx, prompt, []types.Message{{Content: q.Content}})
	if err != nil {
		return "", fmt.Errorf("error from %s: %w", cli, err)
	}

	return response, nil
}
