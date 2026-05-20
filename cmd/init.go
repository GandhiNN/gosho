package cmd

import (
	"fmt"

	"github.com/gandhinn/gosho/config"
	gosso "github.com/gandhinn/gosho/sso"
)

func Init() error {
	fmt.Println("Creating default configuration for gosho...")
	fmt.Println()

	startURL := promptText("SSO start URL")
	region := promptSelect("Region", gosso.Regions)

	config.SaveDefaults(&config.Defaults{
		StartURL: startURL,
		Region:   region,
	})

	fmt.Println("✓ Defaults saved to ~/.gosho/defaults.json")
	return nil
}
