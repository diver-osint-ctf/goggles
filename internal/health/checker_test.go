package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goggles/internal/website"
)

func TestChecker_Check(t *testing.T) {
	tests := []struct {
		name            string
		website         website.Website
		serverResponse  string
		serverStatus    int
		expectedOK      bool
		expectedStatus  int
		expectedHasWord bool
	}{
		{
			name: "successful check with word found",
			website: website.Website{
				URL:        "", // Will be set to test server URL
				CheckWords: []string{"test", "content"},
			},
			serverResponse:  "This is test content for checking",
			serverStatus:    200,
			expectedOK:      true,
			expectedStatus:  200,
			expectedHasWord: true,
		},
		{
			name: "successful check with word not found",
			website: website.Website{
				URL:        "", // Will be set to test server URL
				CheckWords: []string{"missing", "word"},
			},
			serverResponse:  "This is test content for checking",
			serverStatus:    200,
			expectedOK:      false,
			expectedStatus:  200,
			expectedHasWord: false,
		},
		{
			name: "successful check with no check words",
			website: website.Website{
				URL:        "", // Will be set to test server URL
				CheckWords: []string{},
			},
			serverResponse:  "Any content",
			serverStatus:    200,
			expectedOK:      true,
			expectedStatus:  200,
			expectedHasWord: true,
		},
		{
			name: "server error",
			website: website.Website{
				URL:        "", // Will be set to test server URL
				CheckWords: []string{"Server"},
			},
			serverResponse:  "Server Error",
			serverStatus:    500,
			expectedOK:      false,
			expectedStatus:  500,
			expectedHasWord: true, // Word found but status code invalid
		},
		{
			name: "not found error",
			website: website.Website{
				URL:        "", // Will be set to test server URL
				CheckWords: []string{"test"},
			},
			serverResponse:  "Not Found",
			serverStatus:    404,
			expectedOK:      false,
			expectedStatus:  404,
			expectedHasWord: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check that proper headers are set
				userAgent := r.Header.Get("User-Agent")
				if userAgent == "" {
					t.Error("User-Agent header not set")
				}
				if !contains(userAgent, "Chrome/135") {
					t.Error("User-Agent doesn't look like a modern Chrome browser")
				}

				// Check for important browser headers
				if r.Header.Get("Sec-Fetch-Dest") == "" {
					t.Error("Sec-Fetch-Dest header not set")
				}
				if r.Header.Get("Sec-Ch-Ua") == "" {
					t.Error("Sec-Ch-Ua header not set")
				}
				if r.Header.Get("Accept-Encoding") == "" {
					t.Error("Accept-Encoding header not set")
				}

				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Set the test server URL
			tt.website.URL = server.URL

			// Create checker and perform check
			checker := NewCheckerForTest(false)
			result := checker.Check(tt.website)

			// Verify results
			if result.OK != tt.expectedOK {
				t.Errorf("Expected OK=%v, got OK=%v", tt.expectedOK, result.OK)
			}
			if result.Status != tt.expectedStatus {
				t.Errorf("Expected Status=%d, got Status=%d", tt.expectedStatus, result.Status)
			}
			if result.HasWord != tt.expectedHasWord {
				t.Errorf("Expected HasWord=%v, got HasWord=%v", tt.expectedHasWord, result.HasWord)
			}
			if result.Website.URL != tt.website.URL {
				t.Errorf("Expected Website.URL=%s, got Website.URL=%s", tt.website.URL, result.Website.URL)
			}
			if result.Date.IsZero() {
				t.Error("Expected Date to be set")
			}
			if time.Since(result.Date) > time.Second {
				t.Error("Expected Date to be recent")
			}
		})
	}
}

func TestChecker_CheckInvalidURL(t *testing.T) {
	website := website.Website{
		URL:        "not-a-valid-url",
		CheckWords: []string{"test"},
	}

	checker := NewCheckerForTest(false)
	result := checker.Check(website)

	if result.OK {
		t.Error("Expected OK=false for invalid URL")
	}
	if result.Error == "" {
		t.Error("Expected Error to be set for invalid URL")
	}
	if !contains(result.Error, "request failed") {
		t.Errorf("Expected error about request failure, got: %s", result.Error)
	}
}

func TestChecker_CheckTimeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(32 * time.Second) // Longer than our timeout (30s + 1-3s delay)
		w.WriteHeader(200)
		w.Write([]byte("delayed response"))
	}))
	defer server.Close()

	website := website.Website{
		URL:        server.URL,
		CheckWords: []string{"test"},
	}

	checker := NewCheckerForTest(false)
	start := time.Now()
	result := checker.Check(website)
	elapsed := time.Since(start)

	if result.OK {
		t.Error("Expected OK=false for timeout")
	}
	if result.Error == "" {
		t.Error("Expected Error to be set for timeout")
	}

	// Should timeout around 30 seconds (no delay in test mode)
	if elapsed < 29*time.Second || elapsed > 32*time.Second {
		t.Errorf("Expected timeout around 30 seconds, got %v", elapsed)
	}
}

func TestNewChecker(t *testing.T) {
	checker := NewCheckerForTest(false)

	if checker == nil {
		t.Error("Expected non-nil checker")
	}
	if checker.client == nil {
		t.Error("Expected non-nil HTTP client")
	}
	if checker.client.Timeout != 30*time.Second {
		t.Errorf("Expected timeout of 30s, got %v", checker.client.Timeout)
	}
}

func TestChecker_VerboseMode(t *testing.T) {
	// Test with verbose mode enabled
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("<html><body>Not Found</body></html>"))
	}))
	defer server.Close()

	website := website.Website{
		URL:        server.URL,
		CheckWords: []string{"Success"}, // This word won't be found
	}

	// Test verbose mode
	checkerVerbose := NewCheckerForTest(true)
	result := checkerVerbose.Check(website)

	if result.OK {
		t.Error("Expected OK=false for failed health check")
	}
	if result.HTMLContent == "" {
		t.Error("Expected HTML content to be stored in verbose mode")
	}
	if !contains(result.HTMLContent, "Not Found") {
		t.Error("Expected HTML content to contain response body")
	}

	// Test non-verbose mode
	checkerNonVerbose := NewCheckerForTest(false)
	resultNonVerbose := checkerNonVerbose.Check(website)

	if resultNonVerbose.HTMLContent != "" {
		t.Error("Expected no HTML content to be stored in non-verbose mode")
	}
}

func TestNewCheckerVerboseFlag(t *testing.T) {
	// Test verbose flag is set correctly
	checkerVerbose := NewCheckerForTest(true)
	if !checkerVerbose.verbose {
		t.Error("Expected verbose flag to be true")
	}

	checkerNonVerbose := NewCheckerForTest(false)
	if checkerNonVerbose.verbose {
		t.Error("Expected verbose flag to be false")
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
