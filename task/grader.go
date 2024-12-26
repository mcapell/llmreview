package task

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mcapell/llmreview/llm"
	"github.com/mcapell/llmreview/llm/types"
	"github.com/mcapell/llmreview/question"
)

//go:embed gradingprompt.txt
var gradingPrompt string

type GradeResult struct {
	Grades []struct {
		Grade    int    `json:"grade,omitempty"`
		Category string `json:"category,omitempty"`
		Notes    string `json:"notes,omitempty"`
	} `json:"grades,omitempty"`
}

type Grader struct {
	cli llm.Client
}

func NewGrader() Grader {
	grader, err := llm.New("openai")
	if err != nil {
		panic(err)
	}

	return Grader{
		cli: grader,
	}
}

func (g *Grader) GradeQuestionResult(ctx context.Context, q question.Question, result string) (string, error) {
	response, err := g.cli.Chat(ctx, gradingPrompt+q.Correction, []types.Message{{Content: []types.Content{{Text: result}}}})
	if err != nil {
		return "", fmt.Errorf("error from %s: %w", g.cli, err)
	}

	var grade GradeResult
	if err := json.NewDecoder(strings.NewReader(response)).Decode(&grade); err != nil {
		return "", fmt.Errorf("error parsing grade result: %w", err)
	}

	return response, nil
}
