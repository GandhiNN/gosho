package cmd

import (
	"fmt"

	"github.com/gandhinn/gosho/config"
	gosso "github.com/gandhinn/gosho/sso"
)

func Init() error {
	fmt.Println("Creating gosho configuration...")
	fmt.Println()

	cfg := config.Load()
	cfg.StartURL = promptText("SSO start URL")
	cfg.Region = promptSelect("Region", gosso.Regions)

	if err := config.Save(cfg); err != nil {
		return err
	}

	fmt.Printf("\n✓ Config saved to %s\n", config.Path())
	return nil
}
