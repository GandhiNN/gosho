package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileReturnsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.yaml")
	t.Setenv("GOSHO_CONFIG", path)

	cfg := Load()
	if cfg == nil {
		t.Fatal("Load() returned nil")
	}
	if cfg.StartURL != "" || cfg.Region != "" || len(cfg.Profiles) != 0 {
		t.Errorf("expected empty config, got %+v", cfg)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("GOSHO_CONFIG", path)

	want := &Config{
		StartURL: "https://example.awsapps.com/start",
		Region:   "eu-west-1",
		Profiles: map[string]Profile{
			"aws-dev": {AccountID: "111111111111", Role: "DevOps"},
			"aws-prd": {AccountID: "222222222222", Role: "DevOps"},
		},
	}

	if err := Save(want); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got := Load()
	if got.StartURL != want.StartURL {
		t.Errorf("StartURL = %q, want %q", got.StartURL, want.StartURL)
	}
	if got.Region != want.Region {
		t.Errorf("Region = %q, want %q", got.Region, want.Region)
	}
	if len(got.Profiles) != len(want.Profiles) {
		t.Fatalf("Profiles len = %d, want %d", len(got.Profiles), len(want.Profiles))
	}
	for name, wp := range want.Profiles {
		gp, ok := got.Profiles[name]
		if !ok {
			t.Errorf("missing profile %q", name)
			continue
		}
		if gp != wp {
			t.Errorf("profile %q = %+v, want %+v", name, gp, wp)
		}
	}
}

func TestLoadMalformedReturnsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("GOSHO_CONFIG", path)

	// Invalid YAML (unterminated / bad structure for the Config type)
	if err := os.WriteFile(path, []byte("start_url: [this is not valid: for a string"), 0600); err != nil {
		t.Fatalf("write malformed config: %v", err)
	}

	cfg := Load()
	if cfg == nil {
		t.Fatal("Load() returned nil for malformed config")
	}
	if cfg.StartURL != "" || len(cfg.Profiles) != 0 {
		t.Errorf("expected empty config for malformed input, got %+v", cfg)
	}
}

func TestPathRespectsEnv(t *testing.T) {
	custom := filepath.Join(t.TempDir(), "custom.yaml")
	t.Setenv("GOSHO_CONFIG", custom)
	if got := Path(); got != custom {
		t.Errorf("Path() = %q, want %q", got, custom)
	}
}
