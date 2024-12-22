package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/mcapell/llmreview/llm/openai"
	"github.com/mcapell/llmreview/llm/types"
)

var (
	openaiKey = "OPENAI_API_KEY"
)

type Client interface {
	Chat(ctx context.Context, prompt string, msgs []types.Message) (string, error)
}

func New(model string) (Client, error) {
	switch model {
	case "openai":
		apiKey := os.Getenv(openaiKey)
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
		}
		return openai.NewClient(apiKey), nil
	default:
		return nil, fmt.Errorf("unsupported model: %s", model)
	}
}
