package sso

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type AccessToken struct {
	StartURL     string       `json:"startUrl"`
	Region       string       `json:"region"`
	AccessToken  string       `json:"accessToken"`
	ExpiresAt    time.Time    `json:"expiresAt"`
	RefreshToken string       `json:"refreshToken"`
	Client       DeviceClient `json:"deviceClient"`
}

type DeviceClient struct {
	ClientID     string    `json:"clientId"`
	ClientSecret string    `json:"clientSecret"`
	ExpiresAt    time.Time `json:"registrationExpiresAt"`
}

func (t *AccessToken) IsExpired() bool {
	return t.ExpiresAt.Before(time.Now().UTC())
}

func cacheDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gosho", "cache")
}

func cacheFilePath(profile string) string {
	h := sha1.Sum([]byte(profile))
	return filepath.Join(cacheDir(), fmt.Sprintf("%x.json", h))
}

func LoadCachedToken(profile string) (*AccessToken, error) {
	data, err := os.ReadFile(cacheFilePath(profile))
	if err != nil {
		return nil, err
	}
	var token AccessToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

func SaveToken(profile string, token *AccessToken) error {
	os.MkdirAll(cacheDir(), 0700)
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cacheFilePath(profile), data, 0600)
}
