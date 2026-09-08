package cmd

import (
	"testing"

	gosso "github.com/gandhinn/gosho/sso"
)

func TestParseCredentialBlock(t *testing.T) {
	tests := []struct {
		name        string
		block       string
		wantAKID    string
		wantSecret  string
		wantSession string
	}{
		{
			name: "bash export format",
			block: `export AWS_ACCESS_KEY_ID="AKIAEXAMPLE"
export AWS_SECRET_ACCESS_KEY="secretvalue123"
export AWS_SESSION_TOKEN="sessiontoken456"`,
			wantAKID:    "AKIAEXAMPLE",
			wantSecret:  "secretvalue123",
			wantSession: "sessiontoken456",
		},
		{
			name: "windows cmd set format",
			block: `set AWS_ACCESS_KEY_ID=AKIAEXAMPLE
set AWS_SECRET_ACCESS_KEY=secretvalue123
set AWS_SESSION_TOKEN=sessiontoken456`,
			wantAKID:    "AKIAEXAMPLE",
			wantSecret:  "secretvalue123",
			wantSession: "sessiontoken456",
		},
		{
			name: "powershell env format",
			block: `$Env:AWS_ACCESS_KEY_ID="AKIAEXAMPLE"
$Env:AWS_SECRET_ACCESS_KEY="secretvalue123"
$Env:AWS_SESSION_TOKEN="sessiontoken456"`,
			wantAKID:    "AKIAEXAMPLE",
			wantSecret:  "secretvalue123",
			wantSession: "sessiontoken456",
		},
		{
			name: "credentials ini format",
			block: `[default]
aws_access_key_id = AKIAEXAMPLE
aws_secret_access_key = secretvalue123
aws_session_token = sessiontoken456`,
			wantAKID:    "AKIAEXAMPLE",
			wantSecret:  "secretvalue123",
			wantSession: "sessiontoken456",
		},
		{
			name: "single quotes",
			block: `export AWS_ACCESS_KEY_ID='AKIAEXAMPLE'
export AWS_SECRET_ACCESS_KEY='secretvalue123'
export AWS_SESSION_TOKEN='sessiontoken456'`,
			wantAKID:    "AKIAEXAMPLE",
			wantSecret:  "secretvalue123",
			wantSession: "sessiontoken456",
		},
		{
			name:        "empty block",
			block:       "",
			wantAKID:    "",
			wantSecret:  "",
			wantSession: "",
		},
		{
			name: "session token containing special chars",
			block: `export AWS_ACCESS_KEY_ID="AKIAEXAMPLE"
export AWS_SECRET_ACCESS_KEY="abc/def+ghi=="
export AWS_SESSION_TOKEN="FwoGZXIvYXdz//EXAMPLE+token/value=="`,
			wantAKID:    "AKIAEXAMPLE",
			wantSecret:  "abc/def+ghi==",
			wantSession: "FwoGZXIvYXdz//EXAMPLE+token/value==",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseCredentialBlock(tt.block)
			if got.AccessKeyID != tt.wantAKID {
				t.Errorf("AccessKeyID = %q, want %q", got.AccessKeyID, tt.wantAKID)
			}
			if got.SecretAccessKey != tt.wantSecret {
				t.Errorf("SecretAccessKey = %q, want %q", got.SecretAccessKey, tt.wantSecret)
			}
			if got.SessionToken != tt.wantSession {
				t.Errorf("SessionToken = %q, want %q", got.SessionToken, tt.wantSession)
			}
		})
	}
}

func TestValidateCredentials(t *testing.T) {
	tests := []struct {
		name    string
		akid    string
		secret  string
		session string
		wantErr bool
	}{
		{"all present", "a", "b", "c", false},
		{"missing access key", "", "b", "c", true},
		{"missing secret", "a", "", "c", true},
		{"missing session", "a", "b", "", true},
		{"all missing", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds := &gosso.RoleCredentials{
				AccessKeyID:     tt.akid,
				SecretAccessKey: tt.secret,
				SessionToken:    tt.session,
			}
			_, err := validateCredentials(creds)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCredentials() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
