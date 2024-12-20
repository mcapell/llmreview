package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mcapell/llmreview/llm/types"
)

var (
	defaultTemperature             = 1.0
	defaultMaxCompletionTokens     = 2000 // use a defined upper limit
	defaultNumberOfChatCompletions = 1
)

type ChatRequest struct {
	Model               string        `json:"model,omitempty"`
	Messages            []ChatMessage `json:"messages,omitempty"`
	Temperature         float64       `json:"temperature,omitempty"`
	MaxCompletionTokens int           `json:"max_completion_tokens,omitempty"`
	NumberOfChats       int           `json:"n,omitempty"`
}

type ChatMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type ChatResponse struct {
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Content string `json:"content"`
			Role    string `json:"role"`
		} `json:"message,omitempty"`
	} `json:"choices,omitempty"`
}

type Client struct {
	model       string
	apiKey      string
	temperature float64
}

func NewClient(apiKey string) *Client {
	return &Client{
		model:       "gpt-4o-mini",
		apiKey:      apiKey,
		temperature: defaultTemperature,
	}
}

func (c *Client) String() string {
	return fmt.Sprintf("%s - temperature: %0.1f", c.model, c.temperature)
}

func (c *Client) Chat(ctx context.Context, msg types.Message) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	messages := []ChatMessage{
		{Content: msg.Text, Role: "user"},
	}

	if msg.Prompt != "" {
		messages = append(messages, ChatMessage{Content: msg.Prompt, Role: "system"})
	}

	payload := ChatRequest{
		Model:               c.model,
		Messages:            messages,
		Temperature:         c.temperature,
		MaxCompletionTokens: defaultMaxCompletionTokens,
		NumberOfChats:       defaultNumberOfChatCompletions,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error from API: %s", string(bodyBytes))
	}

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	finishReason := result.Choices[0].FinishReason
	if finishReason != "stop" {
		return "", fmt.Errorf("unexpected finish reason: %s", finishReason)
	}

	return "", fmt.Errorf("no choices returned from API")
}
