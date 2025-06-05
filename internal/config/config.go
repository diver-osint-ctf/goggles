package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	DiscordWebhookURL string   `yaml:"discord_webhook_url"`
	Timezone          string   `yaml:"timezone"`
	Start             string   `yaml:"start"`
	End               string   `yaml:"end"`
	Mentions          []string `yaml:"mentions"`
	Threads           int      `yaml:"threads"`
}

// Load reads and parses the configuration file
func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}
