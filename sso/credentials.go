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
	var err error
	if _, err = os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(path), 0700)
		cfg = ini.Empty()
	} else {
		cfg, err = ini.Load(path)
		if err != nil {
			return fmt.Errorf("load credentials file: %w", err)
		}
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

func credentialsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aws", "credentials")
}
