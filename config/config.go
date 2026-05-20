package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	AccountID string `yaml:"account_id"`
	Role      string `yaml:"role"`
}

type Config struct {
	StartURL string             `yaml:"start_url"`
	Region   string             `yaml:"region"`
	Profiles map[string]Profile `yaml:"profiles"`
}

func Path() string {
	if p := os.Getenv("GOSHO_CONFIG"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gosho", "config.yaml")
}

func Load() *Config {
	data, err := os.ReadFile(Path())
	if err != nil {
		return &Config{}
	}
	var cfg Config
	yaml.Unmarshal(data, &cfg)
	return &cfg
}

func Save(cfg *Config) error {
	path := Path()
	os.MkdirAll(filepath.Dir(path), 0700)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
