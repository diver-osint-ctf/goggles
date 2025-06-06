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
	const batchSize = 5

	// If failures are 5 or less, send as usual
	if len(failures) <= batchSize {
		report := c.generateReport(failures)
		return c.sendSingleReport(report, mentions)
	}

	// Split failures into batches of 5
	for i := 0; i < len(failures); i += batchSize {
		end := i + batchSize
		if end > len(failures) {
			end = len(failures)
		}

		batch := failures[i:end]
		batchNumber := (i / batchSize) + 1
		totalBatches := (len(failures) + batchSize - 1) / batchSize

		report := c.generateBatchReport(batch, batchNumber, totalBatches)

		err := c.sendSingleReport(report, mentions)
		if err != nil {
			return fmt.Errorf("failed to send batch %d: %w", batchNumber, err)
		}
	}

	return nil
}

// sendSingleReport sends a single report message
func (c *Client) sendSingleReport(report string, mentions []string) error {
	content := ""
	if len(mentions) > 0 {
		mentionStrs := make([]string, len(mentions))
		for i, mention := range mentions {
			mentionStrs[i] = fmt.Sprintf("<@%s>", mention)
		}
		content = strings.Join(mentionStrs, " ")
	}

	message := Message{
		Content: content + "\n```markdown\n" + report + "\n```",
		TTS:     false,
	}

	return c.sendMessage(message)
}

// generateBatchReport creates a markdown report of failures for a specific batch
func (c *Client) generateBatchReport(failures []health.Result, batchNumber, totalBatches int) string {
	var sb strings.Builder

	if totalBatches > 1 {
		sb.WriteString(fmt.Sprintf("# Failed Websites Report (%d/%d)\n\n", batchNumber, totalBatches))
	} else {
		sb.WriteString("# Failed Websites Report\n\n")
	}

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
