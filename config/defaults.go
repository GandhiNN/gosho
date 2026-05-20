package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Defaults struct {
	StartURL string `json:"start_url"`
	Region   string `json:"region"`
}

func defaultsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gosho", "defaults.json")
}

func LoadDefaults() *Defaults {
	data, err := os.ReadFile(defaultsPath())
	if err != nil {
		return &Defaults{}
	}
	var d Defaults
	json.Unmarshal(data, &d)
	return &d
}

func SaveDefaults(d *Defaults) {
	dir := filepath.Dir(defaultsPath())
	os.MkdirAll(dir, 0700)
	data, _ := json.MarshalIndent(d, "", "	")
	os.WriteFile(defaultsPath(), data, 0600)
}
