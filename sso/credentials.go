package sso

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

func WriteCredentials(profile string, creds *RoleCredentials, region string) error {
	path := credentialsPath()

	var cfg *ini.File
	if _, statErr := os.Stat(path); statErr != nil {
		if !os.IsNotExist(statErr) {
			return fmt.Errorf("stat credentials file: %w", statErr)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return fmt.Errorf("create credentials dir: %w", err)
		}
		cfg = ini.Empty()
	} else {
		loaded, err := ini.Load(path)
		if err != nil {
			return fmt.Errorf("load credentials file: %w", err)
		}
		cfg = loaded
	}

	sec, _ := cfg.NewSection(profile)
	sec.Key("aws_access_key_id").SetValue(creds.AccessKeyID)
	sec.Key("aws_secret_access_key").SetValue(creds.SecretAccessKey)
	sec.Key("aws_session_token").SetValue(creds.SessionToken)
	if region != "" {
		sec.Key("region").SetValue(region)
	}

	return cfg.SaveTo(path)
}

func RemoveCredentials(profile string) error {
	path := credentialsPath()
	cfg, err := ini.Load(path)
	if err != nil {
		return nil // file does not exist
	}
	cfg.DeleteSection(profile)
	return cfg.SaveTo(path)
}

// ListCredentialProfiles returns the profile (section) names present in the
// AWS shared credentials file, excluding the ini default section. Returns nil
// if the file does not exist.
func ListCredentialProfiles() []string {
	cfg, err := ini.Load(credentialsPath())
	if err != nil {
		return nil
	}
	var profiles []string
	for _, sec := range cfg.Sections() {
		if sec.Name() == ini.DefaultSection {
			continue
		}
		profiles = append(profiles, sec.Name())
	}
	return profiles
}

func credentialsPath() string {
	if p := os.Getenv("AWS_SHARED_CREDENTIALS_FILE"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aws", "credentials")
}
