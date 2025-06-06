package discord

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goggles/internal/health"
	"goggles/internal/website"
)

func TestClient_SendSummary(t *testing.T) {
	tests := []struct {
		name       string
		failures   []health.Result
		totalCount int
		wantTitle  string
		wantColor  int
	}{
		{
			name:       "no failures",
			failures:   []health.Result{},
			totalCount: 3,
			wantTitle:  "Nice! 3 All websites are healthy :tada:",
			wantColor:  successColor,
		},
		{
			name: "with failures",
			failures: []health.Result{
				{
					Website: website.Website{URL: "https://example.com"},
					OK:      false,
				},
			},
			totalCount: 3,
			wantTitle:  "1 Website(s) are unhealthy :tired_face:",
			wantColor:  failureColor,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server to capture webhook request
			var receivedMessage Message
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}

				contentType := r.Header.Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected application/json content type, got %s", contentType)
				}

				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("Failed to read request body: %v", err)
				}

				err = json.Unmarshal(body, &receivedMessage)
				if err != nil {
					t.Fatalf("Failed to unmarshal request body: %v", err)
				}

				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			// Test the client
			client := NewClient(server.URL)
			err := client.SendSummary(tt.failures, tt.totalCount)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify the message content
			if len(receivedMessage.Embeds) != 1 {
				t.Errorf("Expected 1 embed, got %d", len(receivedMessage.Embeds))
			}

			embed := receivedMessage.Embeds[0]
			if embed.Title != tt.wantTitle {
				t.Errorf("Expected title '%s', got '%s'", tt.wantTitle, embed.Title)
			}

			if embed.Color != tt.wantColor {
				t.Errorf("Expected color %d, got %d", tt.wantColor, embed.Color)
			}

			if receivedMessage.TTS != false {
				t.Error("Expected TTS to be false")
			}
		})
	}
}

func TestClient_SendReport(t *testing.T) {
	failures := []health.Result{
		{
			Website: website.Website{
				URL:        "https://example.com",
				CheckWords: []string{"test", "word"},
			},
			OK:      false,
			Status:  404,
			HasWord: false,
			Error:   "Not found",
		},
	}
	mentions := []string{"123456789", "987654321"}

	// Create test server to capture webhook request
	var receivedMessage Message
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		err = json.Unmarshal(body, &receivedMessage)
		if err != nil {
			t.Fatalf("Failed to unmarshal request body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test the client
	client := NewClient(server.URL)
	err := client.SendReport(failures, mentions)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify mentions are included
	expectedContent := "<@123456789> <@987654321>"
	if !contains(receivedMessage.Content, expectedContent) {
		t.Errorf("Expected content to contain '%s', got '%s'", expectedContent, receivedMessage.Content)
	}

	// Verify markdown report is included
	if !contains(receivedMessage.Content, "```markdown") {
		t.Error("Expected markdown code block in content")
	}
	if !contains(receivedMessage.Content, "https://example.com") {
		t.Error("Expected URL in report content")
	}
}

func TestClient_SendSummaryHTTPError(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.SendSummary([]health.Result{}, 1)

	if err == nil {
		t.Error("Expected error for HTTP 400 response")
	}
	if !contains(err.Error(), "Discord API returned status: 400") {
		t.Errorf("Expected error about Discord API status, got: %v", err)
	}
}

func TestClient_generateReport(t *testing.T) {
	failures := []health.Result{
		{
			Website: website.Website{
				URL:        "https://example.com",
				CheckWords: []string{"test", "word"},
			},
			OK:      false,
			Status:  404,
			HasWord: false,
			Date:    time.Now(),
			Error:   "Not found",
		},
		{
			Website: website.Website{
				URL:        "https://another.com",
				CheckWords: []string{},
			},
			OK:      false,
			Status:  500,
			HasWord: true,
			Date:    time.Now(),
		},
	}

	client := NewClient("https://example.com/webhook")
	report := client.generateReport(failures)

	// Verify report contains expected elements
	expectedElements := []string{
		"# Failed Websites Report",
		"- url: https://example.com",
		"- url: https://another.com",
		"- check words:",
		"- test",
		"- word",
		"- factors:",
		"- HTTP Response Code: 404",
		"- HTTP Response Code: 500",
		"- Has Check Words: false",
		"- Has Check Words: true",
		"- Error: Not found",
		"Reported by [Goggles]",
	}

	for _, element := range expectedElements {
		if !contains(report, element) {
			t.Errorf("Expected report to contain '%s', got:\n%s", element, report)
		}
	}
}

func TestClient_SendReportBatching(t *testing.T) {
	// Create 12 failures to test batching (should result in 3 batches of 5, 5, and 2)
	failures := make([]health.Result, 12)
	for i := 0; i < 12; i++ {
		failures[i] = health.Result{
			Website: website.Website{
				URL:        fmt.Sprintf("https://example%d.com", i+1),
				CheckWords: []string{"test"},
			},
			OK:      false,
			Status:  404,
			HasWord: false,
			Error:   "Not found",
		}
	}

	mentions := []string{"123456789"}

	// Track the number of requests received
	requestCount := 0
	var receivedMessages []Message

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		var message Message
		err = json.Unmarshal(body, &message)
		if err != nil {
			t.Fatalf("Failed to unmarshal request body: %v", err)
		}

		receivedMessages = append(receivedMessages, message)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test the client
	client := NewClient(server.URL)
	err := client.SendReport(failures, mentions)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should receive 3 batched requests (5 + 5 + 2 = 12 failures)
	expectedRequests := 3
	if requestCount != expectedRequests {
		t.Errorf("Expected %d requests, got %d", expectedRequests, requestCount)
	}

	// Verify each batch contains the correct batch information
	for i, message := range receivedMessages {
		batchNumber := i + 1
		expectedBatchHeader := fmt.Sprintf("# Failed Websites Report (%d/3)", batchNumber)

		if !contains(message.Content, expectedBatchHeader) {
			t.Errorf("Expected batch %d to contain header '%s'", batchNumber, expectedBatchHeader)
		}

		// Verify mentions are included in each batch
		if !contains(message.Content, "<@123456789>") {
			t.Errorf("Expected mentions in batch %d", batchNumber)
		}
	}
}

func TestClient_SendReportNoBatching(t *testing.T) {
	// Create only 3 failures (should not be batched)
	failures := make([]health.Result, 3)
	for i := 0; i < 3; i++ {
		failures[i] = health.Result{
			Website: website.Website{
				URL:        fmt.Sprintf("https://example%d.com", i+1),
				CheckWords: []string{"test"},
			},
			OK:      false,
			Status:  404,
			HasWord: false,
		}
	}

	requestCount := 0
	var receivedMessage Message

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		err = json.Unmarshal(body, &receivedMessage)
		if err != nil {
			t.Fatalf("Failed to unmarshal request body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test the client
	client := NewClient(server.URL)
	err := client.SendReport(failures, []string{})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should receive only 1 request (no batching needed)
	if requestCount != 1 {
		t.Errorf("Expected 1 request, got %d", requestCount)
	}

	// Should not contain batch information
	if contains(receivedMessage.Content, "1/1") {
		t.Error("Expected no batch information for small report")
	}
}

func TestNewClient(t *testing.T) {
	webhookURL := "https://discord.com/api/webhooks/test"
	client := NewClient(webhookURL)

	if client == nil {
		t.Error("Expected non-nil client")
	}
	if client.webhookURL != webhookURL {
		t.Errorf("Expected webhook URL '%s', got '%s'", webhookURL, client.webhookURL)
	}
	if client.httpClient == nil {
		t.Error("Expected non-nil HTTP client")
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
