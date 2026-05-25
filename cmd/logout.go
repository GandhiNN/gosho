package cmd

import (
	"fmt"

	gosso "github.com/gandhinn/gosho/sso"
)

func Logout(profile string) error {
	if profile == "all" {
		profiles := gosso.ListProfiles()
		if len(profiles) == 0 {
			fmt.Println("No cached profiles.")
			return nil
		}
		for _, p := range profiles {
			gosso.RemoveToken(p)
			gosso.RemoveCredentials(p)
			fmt.Printf("✓ Logged out: %s\n", p)
		}
		return nil
	}

	if profile == "" {
		profile = promptText("Profile to logout")
	}

	gosso.RemoveToken(profile)
	gosso.RemoveCredentials(profile)

	fmt.Printf("✓ Logged out: %s (token + credentials removed)\n", profile)
	return nil
}
