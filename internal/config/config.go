package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultModel         = "llama-3.3-70b-versatile"
	DefaultMaxIterations = 25
	DefaultTimeout       = 60
)

type Config struct {
	APIKey             string `yaml:"api_key"`
	Model              string `yaml:"model"`
	MaxIterations      int    `yaml:"max_iterations"`
	AutoApproveReads   bool   `yaml:"auto_approve_reads"`
	CustomSystemPrompt string `yaml:"custom_system_prompt"`
	Theme              string `yaml:"theme"`
}

func DefaultConfig() *Config {
	return &Config{
		Model:            DefaultModel,
		MaxIterations:    DefaultMaxIterations,
		AutoApproveReads: true,
		Theme:            "dark",
	}
}

func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Load from config file
	configPath := filepath.Join(configDir(), "config.yaml")
	data, err := os.ReadFile(configPath)
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config: %w", err)
		}
	}

	// Environment variable overrides
	if key := os.Getenv("GROQ_API_KEY"); key != "" {
		cfg.APIKey = key
	}

	if cfg.Model == "" {
		cfg.Model = DefaultModel
	}
	if cfg.MaxIterations == 0 {
		cfg.MaxIterations = DefaultMaxIterations
	}

	return cfg, nil
}

func Save(cfg *Config) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, "config.yaml"), data, 0o600)
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".groqcode")
}

func ConfigDir() string {
	return configDir()
}
