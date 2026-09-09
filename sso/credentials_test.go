package sso

import (
	"path/filepath"
	"testing"

	"gopkg.in/ini.v1"
)

func TestWriteCredentialsCreatesFileAndSection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", path)

	creds := &RoleCredentials{
		AccessKeyID:     "AKIAEXAMPLE",
		SecretAccessKey: "secret",
		SessionToken:    "token",
	}
	if err := WriteCredentials("aws-dev", creds, "eu-west-1"); err != nil {
		t.Fatalf("WriteCredentials() error: %v", err)
	}

	cfg, err := ini.Load(path)
	if err != nil {
		t.Fatalf("load written file: %v", err)
	}
	sec := cfg.Section("aws-dev")
	if got := sec.Key("aws_access_key_id").String(); got != "AKIAEXAMPLE" {
		t.Errorf("aws_access_key_id = %q, want AKIAEXAMPLE", got)
	}
	if got := sec.Key("aws_secret_access_key").String(); got != "secret" {
		t.Errorf("aws_secret_access_key = %q, want secret", got)
	}
	if got := sec.Key("aws_session_token").String(); got != "token" {
		t.Errorf("aws_session_token = %q, want token", got)
	}
	if got := sec.Key("region").String(); got != "eu-west-1" {
		t.Errorf("region = %q, want eu-west-1", got)
	}
}

func TestWriteCredentialsPreservesOtherProfiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", path)

	if err := WriteCredentials("aws-dev", &RoleCredentials{
		AccessKeyID: "AKIADEV", SecretAccessKey: "s1", SessionToken: "t1",
	}, "eu-west-1"); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := WriteCredentials("aws-prd", &RoleCredentials{
		AccessKeyID: "AKIAPRD", SecretAccessKey: "s2", SessionToken: "t2",
	}, "us-east-1"); err != nil {
		t.Fatalf("second write: %v", err)
	}

	cfg, err := ini.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got := cfg.Section("aws-dev").Key("aws_access_key_id").String(); got != "AKIADEV" {
		t.Errorf("aws-dev key = %q, want AKIADEV (first profile not preserved)", got)
	}
	if got := cfg.Section("aws-prd").Key("aws_access_key_id").String(); got != "AKIAPRD" {
		t.Errorf("aws-prd key = %q, want AKIAPRD", got)
	}
}

func TestWriteCredentialsNoRegion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", path)

	if err := WriteCredentials("aws-dev", &RoleCredentials{
		AccessKeyID: "AKIA", SecretAccessKey: "s", SessionToken: "t",
	}, ""); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg, _ := ini.Load(path)
	if cfg.Section("aws-dev").HasKey("region") {
		t.Error("region key should not be set when region is empty")
	}
}

func TestRemoveCredentials(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", path)

	if err := WriteCredentials("aws-dev", &RoleCredentials{
		AccessKeyID: "AKIADEV", SecretAccessKey: "s1", SessionToken: "t1",
	}, "eu-west-1"); err != nil {
		t.Fatalf("write dev: %v", err)
	}
	if err := WriteCredentials("aws-prd", &RoleCredentials{
		AccessKeyID: "AKIAPRD", SecretAccessKey: "s2", SessionToken: "t2",
	}, "us-east-1"); err != nil {
		t.Fatalf("write prd: %v", err)
	}

	if err := RemoveCredentials("aws-dev"); err != nil {
		t.Fatalf("RemoveCredentials: %v", err)
	}

	cfg, _ := ini.Load(path)
	if cfg.HasSection("aws-dev") {
		t.Error("aws-dev section should have been removed")
	}
	if !cfg.HasSection("aws-prd") {
		t.Error("aws-prd section should still exist")
	}
}

func TestRemoveCredentialsMissingFileNoError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", path)

	if err := RemoveCredentials("nonexistent"); err != nil {
		t.Errorf("RemoveCredentials on missing file returned error: %v", err)
	}
}

func TestCredentialsPathRespectsEnv(t *testing.T) {
	custom := filepath.Join(t.TempDir(), "custom-creds")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", custom)
	if got := credentialsPath(); got != custom {
		t.Errorf("credentialsPath() = %q, want %q", got, custom)
	}
}
