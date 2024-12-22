package question

import (
	"strings"

	"github.com/mcapell/llmreview/llm/types"
)

type InputType string

const (
	InputText InputType = "text"
	InputPDF  InputType = "pdf"
)

type Question struct {
	Path           string
	QuestionType   InputType
	CorrectionType InputType
	Content        []types.Content
	Correction     string
}

func (q *Question) QuestionPath() string {
	return q.Path
}

func (q *Question) CorrectionPath() string {
	return addPathSuffix(q.Path, correctionSuffix)
}

func (q *Question) GradingPath() string {
	return changeExtension(addPathSuffix(q.Path, gradeSuffix), "json")
}

func addPathSuffix(path, suffix string) string {
	if lastN := strings.LastIndex(path, "."); lastN != -1 {
		return path[:lastN] + suffix + path[lastN:]
	}
	return path
}

func changeExtension(path, extension string) string {
	if lastN := strings.LastIndex(path, "."); lastN != -1 {
		return path[:lastN] + "." + extension
	}
	return path
}
