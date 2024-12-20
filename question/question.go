package question

type Question struct {
	Path           string
	QuestionType   InputType
	CorrectionType InputType
	Text           string
	Correction     string
}

type InputType string

const (
	InputText InputType = "text"
)
