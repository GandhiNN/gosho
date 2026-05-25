package cmd

import (
	"fmt"

	gosso "github.com/gandhinn/gosho/sso"
)

func Logout(profile string) error {
	if profile == "" {
		profile = promptText("Profile to logout")
	}

	gosso.RemoveToken(profile)
	gosso.RemoveCredentials(profile)

	fmt.Printf("✓ Logged out: %s (token + credentials removed)\n", profile)
	return nil
}
