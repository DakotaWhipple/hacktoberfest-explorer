package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds application configuration
type Config struct {
	GitHubToken        string   `json:"github_token"`
	PreferredLanguages []string `json:"preferred_languages"`
	SkillLevel         string   `json:"skill_level"` // beginner, intermediate, advanced
	MaxRepos           int      `json:"max_repos"`
	MaxIssuesPerRepo   int      `json:"max_issues_per_repo"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		PreferredLanguages: []string{"Go", "JavaScript", "Python", "TypeScript"},
		SkillLevel:         "intermediate",
		MaxRepos:           50,
		MaxIssuesPerRepo:   20,
	}
}

// Load configuration from environment and config file
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Try to load from config file
	if err := loadFromFile(cfg); err != nil {
		// Config file is optional, continue with defaults
	}

	// Override with environment variables
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		cfg.GitHubToken = token
	}

	return cfg, nil
}

// loadFromFile attempts to load configuration from ~/.hacktober-config.json
func loadFromFile(cfg *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, ".hacktober-config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, cfg)
}

// Save configuration to file
func (c *Config) Save() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, ".hacktober-config.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
