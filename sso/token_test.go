package sso

import (
	"testing"
	"time"
)

// isolateHome points os.UserHomeDir() at a temp dir for the duration of a test.
// os.UserHomeDir reads HOME on Unix and USERPROFILE on Windows, so set both.
func isolateHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return home
}

func TestIsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{"future", time.Now().UTC().Add(time.Hour), false},
		{"past", time.Now().UTC().Add(-time.Hour), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok := &AccessToken{ExpiresAt: tt.expiresAt}
			if got := tok.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacheFilePathDeterministicAndUnique(t *testing.T) {
	isolateHome(t)

	a1 := cacheFilePath("aws-dev")
	a2 := cacheFilePath("aws-dev")
	b := cacheFilePath("aws-prd")

	if a1 != a2 {
		t.Errorf("cacheFilePath not deterministic: %q vs %q", a1, a2)
	}
	if a1 == b {
		t.Error("cacheFilePath should differ for different profiles")
	}
}

func TestSaveLoadTokenRoundTrip(t *testing.T) {
	isolateHome(t)

	want := &AccessToken{
		StartURL:     "https://example.awsapps.com/start",
		Region:       "eu-west-1",
		AccessToken:  "access-token-value",
		ExpiresAt:    time.Now().UTC().Add(time.Hour).Truncate(time.Second),
		RefreshToken: "refresh-token-value",
		Client: DeviceClient{
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			ExpiresAt:    time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second),
		},
	}

	if err := SaveToken("aws-dev", want); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}

	got, err := LoadCachedToken("aws-dev")
	if err != nil {
		t.Fatalf("LoadCachedToken: %v", err)
	}
	if got.StartURL != want.StartURL || got.Region != want.Region ||
		got.AccessToken != want.AccessToken || got.RefreshToken != want.RefreshToken {
		t.Errorf("round-trip mismatch:\n got %+v\n want %+v", got, want)
	}
	if !got.ExpiresAt.Equal(want.ExpiresAt) {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, want.ExpiresAt)
	}
	if got.Client.ClientID != want.Client.ClientID {
		t.Errorf("Client.ClientID = %q, want %q", got.Client.ClientID, want.Client.ClientID)
	}
}

func TestLoadCachedTokenMissing(t *testing.T) {
	isolateHome(t)

	if _, err := LoadCachedToken("nonexistent"); err == nil {
		t.Error("expected error loading nonexistent token, got nil")
	}
}

func TestSaveTokenUpdatesIndex(t *testing.T) {
	isolateHome(t)

	tok := &AccessToken{ExpiresAt: time.Now().UTC().Add(time.Hour)}
	if err := SaveToken("aws-dev", tok); err != nil {
		t.Fatalf("SaveToken dev: %v", err)
	}
	if err := SaveToken("aws-prd", tok); err != nil {
		t.Fatalf("SaveToken prd: %v", err)
	}
	// Saving the same profile again should not duplicate it.
	if err := SaveToken("aws-dev", tok); err != nil {
		t.Fatalf("SaveToken dev again: %v", err)
	}

	profiles := ListProfiles()
	if len(profiles) != 2 {
		t.Fatalf("ListProfiles len = %d, want 2 (%v)", len(profiles), profiles)
	}
	if !contains(profiles, "aws-dev") || !contains(profiles, "aws-prd") {
		t.Errorf("index missing expected profiles: %v", profiles)
	}
}

func TestRemoveTokenUpdatesIndexAndFile(t *testing.T) {
	isolateHome(t)

	tok := &AccessToken{ExpiresAt: time.Now().UTC().Add(time.Hour)}
	if err := SaveToken("aws-dev", tok); err != nil {
		t.Fatalf("SaveToken dev: %v", err)
	}
	if err := SaveToken("aws-prd", tok); err != nil {
		t.Fatalf("SaveToken prd: %v", err)
	}

	if err := RemoveToken("aws-dev"); err != nil {
		t.Fatalf("RemoveToken: %v", err)
	}

	profiles := ListProfiles()
	if contains(profiles, "aws-dev") {
		t.Errorf("aws-dev should have been removed from index: %v", profiles)
	}
	if !contains(profiles, "aws-prd") {
		t.Errorf("aws-prd should remain in index: %v", profiles)
	}
	if _, err := LoadCachedToken("aws-dev"); err == nil {
		t.Error("aws-dev token file should have been removed")
	}
}

func TestRemoveTokenMissingNoError(t *testing.T) {
	isolateHome(t)

	if err := RemoveToken("nonexistent"); err != nil {
		t.Errorf("RemoveToken on missing profile returned error: %v", err)
	}
}

func TestListProfilesEmpty(t *testing.T) {
	isolateHome(t)

	if profiles := ListProfiles(); len(profiles) != 0 {
		t.Errorf("expected no profiles, got %v", profiles)
	}
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
