package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected *Config
		wantErr  bool
	}{
		{
			name: "valid config",
			content: `discord_webhook_url: https://discord.com/api/webhooks/test
timezone: Asia/Tokyo
start: 2025/6/3 21:07:25
end: 2025/6/9 21:07:25
mentions:
  - 425905959723073536
  - 123456789012345678`,
			expected: &Config{
				DiscordWebhookURL: "https://discord.com/api/webhooks/test",
				Timezone:          "Asia/Tokyo",
				Start:             "2025/6/3 21:07:25",
				End:               "2025/6/9 21:07:25",
				Mentions:          []string{"425905959723073536", "123456789012345678"},
			},
			wantErr: false,
		},
		{
			name: "minimal config",
			content: `discord_webhook_url: https://example.com
timezone: UTC
start: 2025/1/1 00:00:00
end: 2025/12/31 23:59:59
mentions: []`,
			expected: &Config{
				DiscordWebhookURL: "https://example.com",
				Timezone:          "UTC",
				Start:             "2025/1/1 00:00:00",
				End:               "2025/12/31 23:59:59",
				Mentions:          []string{},
			},
			wantErr: false,
		},
		{
			name:     "invalid yaml",
			content:  "invalid: yaml: content: [",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpFile, err := os.CreateTemp("", "config_test_*.yaml")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			// Write test content
			_, err = tmpFile.WriteString(tt.content)
			require.NoError(t, err)
			tmpFile.Close()

			// Test Load function
			result, err := Load(tmpFile.Name())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	result, err := Load("nonexistent.yaml")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to read config file")
}
