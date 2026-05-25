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
	if err := os.WriteFile(cacheFilePath(profile), data, 0600); err != nil {
		return err
	}
	// Update profile index
	return addToIndex(profile)
}

func indexPath() string {
	return filepath.Join(cacheDir(), "profiles.json")
}

func addToIndex(profile string) error {
	profiles := ListProfiles()
	for _, p := range profiles {
		if p == profile {
			return nil
		}
	}
	profiles = append(profiles, profile)
	data, _ := json.MarshalIndent(profiles, "", "	")
	return os.WriteFile(indexPath(), data, 0600)
}

func ListProfiles() []string {
	data, err := os.ReadFile(indexPath())
	if err != nil {
		return nil
	}
	var profiles []string
	json.Unmarshal(data, &profiles)
	return profiles
}

func RemoveToken(profile string) error {
	path := cacheFilePath(profile)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return removeFromIndex(profile)
}

func removeFromIndex(profile string) error {
	profiles := ListProfiles()
	filtered := make([]string, 0, len(profiles))
	for _, p := range profiles {
		if p != profile {
			filtered = append(filtered, p)
		}
	}
	data, _ := json.MarshalIndent(filtered, "", "\t")
	return os.WriteFile(indexPath(), data, 0600)
}
