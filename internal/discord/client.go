package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"goggles/internal/health"
)

const (
	successColor = 3066993  // #2ECC71
	failureColor = 15158332 // #E74C3C
)

// Client handles Discord webhook notifications
type Client struct {
	webhookURL string
	httpClient *http.Client
}

// NewClient creates a new Discord client
func NewClient(webhookURL string) *Client {
	return &Client{
		webhookURL: webhookURL,
		httpClient: &http.Client{},
	}
}

// Embed represents a Discord embed
type Embed struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Color       int     `json:"color"`
	Author      *Author `json:"author,omitempty"`
}

// Author represents the author field in Discord embed
type Author struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Message represents a Discord webhook message
type Message struct {
	Content string  `json:"content,omitempty"`
	Embeds  []Embed `json:"embeds,omitempty"`
	TTS     bool    `json:"tts"`
}

// SendSummary sends a summary of health check results
func (c *Client) SendSummary(failures []health.Result, totalCount int) error {
	var embeds []Embed

	if len(failures) == 0 {
		embeds = []Embed{
			{
				Title:       fmt.Sprintf("Nice! %d All websites are healthy :tada:", totalCount),
				Description: "All monitored websites are responding correctly.",
				Color:       successColor,
			},
		}
	} else {
		embeds = []Embed{
			{
				Title:       fmt.Sprintf("%d Website(s) are unhealthy :tired_face:", len(failures)),
				Description: fmt.Sprintf("%d out of %d websites are not responding correctly.", len(failures), totalCount),
				Color:       failureColor,
				Author: &Author{
					Name: "Goggles",
					URL:  "https://github.com/diver-osint-ctf/goggles",
				},
			},
		}
	}

	message := Message{
		Embeds: embeds,
		TTS:    false,
	}

	return c.sendMessage(message)
}

// SendReport sends a detailed report of failures
func (c *Client) SendReport(failures []health.Result, mentions []string) error {
	report := c.generateReport(failures)

	content := ""
	if len(mentions) > 0 {
		mentionStrs := make([]string, len(mentions))
		for i, mention := range mentions {
			mentionStrs[i] = fmt.Sprintf("<@%s>", mention)
		}
		content = strings.Join(mentionStrs, " ")
	}

	// Send as file attachment (simplified version - just send as message content)
	message := Message{
		Content: content + "\n```markdown\n" + report + "\n```",
		TTS:     false,
	}

	return c.sendMessage(message)
}

// generateReport creates a markdown report of failures
func (c *Client) generateReport(failures []health.Result) string {
	var sb strings.Builder
	sb.WriteString("# Failed Websites Report\n\n")

	for _, failure := range failures {
		sb.WriteString(fmt.Sprintf("- url: %s\n", failure.Website.URL))

		if len(failure.Website.CheckWords) > 0 {
			sb.WriteString("  - check words:\n")
			for _, word := range failure.Website.CheckWords {
				sb.WriteString(fmt.Sprintf("    - %s\n", word))
			}
		}

		sb.WriteString("  - factors:\n")
		sb.WriteString(fmt.Sprintf("    - HTTP Response Code: %d\n", failure.Status))
		sb.WriteString(fmt.Sprintf("    - Has Check Words: %t\n", failure.HasWord))

		if failure.Error != "" {
			sb.WriteString(fmt.Sprintf("    - Error: %s\n", failure.Error))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Reported by [Goggles](https://github.com/diver-osint-ctf/goggles)")
	return sb.String()
}

// sendMessage sends a message to Discord webhook
func (c *Client) sendMessage(message Message) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequest("POST", c.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Discord API returned status: %d", resp.StatusCode)
	}

	return nil
}
