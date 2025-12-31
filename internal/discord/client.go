package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a Discord webhook client
type Client struct {
	WebhookURL string
	HTTPClient *http.Client
}

// NewClient creates a new Discord webhook client
func NewClient(webhookURL string) *Client {
	return &Client{
		WebhookURL: webhookURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WebhookMessage represents a Discord webhook message
type WebhookMessage struct {
	Content   string  `json:"content,omitempty"`
	Username  string  `json:"username,omitempty"`
	AvatarURL string  `json:"avatar_url,omitempty"`
	Embeds    []Embed `json:"embeds,omitempty"`
}

// Embed represents a Discord embed
type Embed struct {
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	URL         string       `json:"url,omitempty"`
	Color       int          `json:"color,omitempty"`
	Timestamp   string       `json:"timestamp,omitempty"`
	Footer      *EmbedFooter `json:"footer,omitempty"`
	Author      *EmbedAuthor `json:"author,omitempty"`
	Fields      []EmbedField `json:"fields,omitempty"`
}

// EmbedFooter represents a Discord embed footer
type EmbedFooter struct {
	Text    string `json:"text,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

// EmbedAuthor represents a Discord embed author
type EmbedAuthor struct {
	Name    string `json:"name,omitempty"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

// EmbedField represents a Discord embed field
type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// PostSummary posts a summary to Discord
func (c *Client) PostSummary(date string, noteCount int, summary string) error {
	embed := Embed{
		Title:       fmt.Sprintf("📝 %s のMisskeyサマリー", date),
		Description: summary,
		Color:       0x86b300, // Misskey green
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &EmbedFooter{
			Text: fmt.Sprintf("📊 %d件のノートから生成", noteCount),
		},
	}

	msg := WebhookMessage{
		Username: "Misskey Daily Summary",
		Embeds:   []Embed{embed},
	}

	return c.send(msg)
}

// send sends a webhook message to Discord
func (c *Client) send(msg WebhookMessage) error {
	jsonBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequest("POST", c.WebhookURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}
