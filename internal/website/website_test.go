package website

import (
	"os"
	"testing"
)

func TestLoadFromCSV(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expected    []Website
		wantErr     bool
		errContains string
	}{
		{
			name: "valid CSV",
			content: `URL,CHECK_WORDS
https://www.google.com,Google
https://www.yahoo.co.jp,Yahoo`,
			expected: []Website{
				{URL: "https://www.google.com", CheckWords: []string{"Google"}},
				{URL: "https://www.yahoo.co.jp", CheckWords: []string{"Yahoo"}},
			},
			wantErr: false,
		},
		{
			name: "CSV with multiple check words",
			content: `URL,CHECK_WORDS
https://example.com,"word1,word2,word3"`,
			expected: []Website{
				{URL: "https://example.com", CheckWords: []string{"word1", "word2", "word3"}},
			},
			wantErr: false,
		},
		{
			name: "CSV with empty check words",
			content: `URL,CHECK_WORDS
https://example.com,`,
			expected: []Website{
				{URL: "https://example.com", CheckWords: []string{}},
			},
			wantErr: false,
		},
		{
			name: "CSV with empty URL (should skip)",
			content: `URL,CHECK_WORDS
,Google
https://example.com,Test`,
			expected: []Website{
				{URL: "https://example.com", CheckWords: []string{"Test"}},
			},
			wantErr: false,
		},
		{
			name:        "empty CSV",
			content:     "",
			expected:    nil,
			wantErr:     true,
			errContains: "CSV file is empty",
		},
		{
			name: "invalid CSV format",
			content: `URL
https://example.com`,
			expected:    nil,
			wantErr:     true,
			errContains: "expected at least 2 columns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpFile, err := os.CreateTemp("", "website_test_*.csv")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			// Write test content
			_, err = tmpFile.WriteString(tt.content)
			if err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			// Test LoadFromCSV function
			result, err := LoadFromCSV(tmpFile.Name())

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errContains != "" && err != nil {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("Expected error to contain '%s', got: %v", tt.errContains, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !equalWebsites(result, tt.expected) {
					t.Errorf("Expected %+v, got %+v", tt.expected, result)
				}
			}
		})
	}
}

func TestLoadFromCSVNonExistentFile(t *testing.T) {
	result, err := LoadFromCSV("nonexistent.csv")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if result != nil {
		t.Error("Expected nil result for non-existent file")
	}
	if !contains(err.Error(), "failed to open CSV file") {
		t.Errorf("Expected error about opening file, got: %v", err)
	}
}

// Helper functions for testing
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
		findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func equalWebsites(a, b []Website) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].URL != b[i].URL {
			return false
		}
		if !equalStringSlices(a[i].CheckWords, b[i].CheckWords) {
			return false
		}
	}
	return true
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
