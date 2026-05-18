package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	AccountID string `yaml:"account_id"`
	Role      string `yaml:"role"`
}

type Environment struct {
	Profiles map[string]Profile `yaml:"profiles"`
}

type Config struct {
	StartURL string                 `yaml:"start_url"`
	Region   string                 `yaml:"region"`
	Envs     map[string]Environment `yaml:"envs"`
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gosho", "config.yaml")
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}
