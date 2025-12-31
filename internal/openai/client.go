package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultModel   = "gpt-5-mini"
	DefaultBaseURL = "https://api.openai.com/v1"
)

// Client is an OpenAI API client
type Client struct {
	APIKey     string
	Model      string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new OpenAI client
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:  apiKey,
		Model:   DefaultModel,
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// WithModel sets the model to use
func (c *Client) WithModel(model string) *Client {
	c.Model = model
	return c
}

// ChatMessage represents a message in the chat
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents a chat completion request
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

// ChatCompletionResponse represents a chat completion response
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// SummarizeNotes generates a summary of the provided notes
func (c *Client) SummarizeNotes(notes []string, date string) (string, error) {
	if len(notes) == 0 {
		return "この日は投稿がありませんでした。", nil
	}

	// Build the notes content
	notesContent := ""
	for i, note := range notes {
		notesContent += fmt.Sprintf("%d. %s\n", i+1, note)
	}

	systemPrompt := `あなたはソーシャルメディアの投稿を分析して、その日のユーザーの活動をサマリーするアシスタントです。
以下のルールに従ってサマリーを作成してください：
- 日本語で回答してください
- 日記に記載するようなサマリーを想定しています
- ユーザーの気分や活動の傾向を把握してください
- 絵文字の使用は禁止です`

	userPrompt := fmt.Sprintf(`以下は%sのMisskeyへの投稿一覧です。この日の活動をサマリーしてください。

%s`, date, notesContent)

	req := ChatCompletionRequest{
		Model: c.Model,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	resp, err := c.createChatCompletion(req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

// createChatCompletion makes a chat completion API request
func (c *Client) createChatCompletion(req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %d - %s", resp.StatusCode, string(body))
	}

	var chatResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &chatResp, nil
}
