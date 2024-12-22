package question

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/mcapell/llmreview/llm/types"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"golang.org/x/exp/slog"
)

const (
	correctionSuffix = "_correction"
	gradeSuffix      = "_grade"
)

var (
	FileNotFoundErr = errors.New("file not found")
)

func ParseQuestionsFromPath(dataPath string) ([]Question, error) {
	var questions []Question
	filesParsed := map[string]bool{}

	if err := filepath.WalkDir(dataPath, func(path string, d fs.DirEntry, err error) error {
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
			corrPath = changeExtension(addPathSuffix(path, correctionSuffix), "md")
			filesParsed[questionPath] = true
			filesParsed[corrPath] = true
		}

		slog.Debug(fmt.Sprintf("question path: %s - correction path: %s\n", questionPath, corrPath))

		question := Question{
			Path: strings.TrimPrefix(path, dataPath),
		}

		if err := loadQuestionContent(path, &question); err != nil {
			return fmt.Errorf("error parsing question content: %w", err)
		}

		// Correction is optional. Add only if the file exist
		correctionInput, err := parseText(corrPath)
		if err != nil && !errors.Is(err, FileNotFoundErr) {
			return fmt.Errorf("error parsing correction: %w", err)
		} else if err == nil {
			question.Correction = correctionInput[0].Text
			question.CorrectionType = InputText
		}

		questions = append(questions, question)
		return nil
	}); err != nil {
		return nil, err
	}

	return questions, nil
}

func loadQuestionContent(path string, question *Question) error {
	if strings.HasSuffix(path, "md") || strings.HasSuffix(path, "txt") {
		question.QuestionType = InputText
		questionInput, err := parseText(path)
		if err != nil {
			return fmt.Errorf("error parsing text: %w", err)
		}

		question.Content = questionInput
	} else if strings.HasSuffix(path, "pdf") {
		question.QuestionType = InputPDF
		questionInput, err := parsePDF(path)
		if err != nil {
			return fmt.Errorf("error parsing text: %w", err)
		}

		question.Content = questionInput
	} else {
		return fmt.Errorf("unsupported question format")
	}

	return nil
}

func parseText(path string) ([]types.Content, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Join(err, FileNotFoundErr)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return []types.Content{{
		Text: string(content),
	}}, nil
}

// PDF is parsed page-by-page, generating a types.Content object for each one
func parsePDF(path string) ([]types.Content, error) {
	f, doc, err := pdf.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening PDF: %w", err)
	}
	defer f.Close()

	numPages := doc.NumPage()

	var content []types.Content
	// Process each page individually
	for i := range numPages {
		pageContent := types.Content{}

		// Extract images first
		if err := api.ExtractImages(f, []string{fmt.Sprintf("%d", i+1)}, func(img model.Image, _ bool, _ int) error {
			if b, err := io.ReadAll(img.Reader); err != nil {
				pageContent.Images = append(pageContent.Images, b)
			}

			return nil
		}, nil); err != nil {
			return nil, fmt.Errorf("error extracting images: %w", err)
		}

		content = append(content, pageContent)
	}

	// Extract the text
	if _, err := f.Seek(0, 0); err != nil {
		return nil, err
	}

	for i := range numPages {
		page := doc.Page(i + 1)
		text := strings.Builder{}
		for _, t := range page.Content().Text {
			text.WriteString(t.S)
		}

		content[i].Text = text.String()
	}

	return content, nil
}
