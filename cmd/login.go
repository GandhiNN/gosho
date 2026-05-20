package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/gandhinn/gosho/config"
	gosso "github.com/gandhinn/gosho/sso"
	"github.com/manifoldco/promptui"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func Login() error {
	ctx := context.Background()

	// Use saved defaults or prompt
	defaults := config.LoadDefaults()
	startURL := defaults.StartURL
	region := defaults.Region
	if startURL == "" {
		startURL = promptText("SSO start URL")
	} else {
		fmt.Printf("Using start URL: %s\n", startURL)
	}
	if region == "" {
		region = promptSelect("Region", gosso.Regions)
	} else {
		fmt.Printf("Using region: %s\n", region)
	}

	// Build AWS config
	cfg, _ := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))

	// Register OIDC client
	oidcClient := ssooidc.NewFromConfig(cfg)
	fmt.Print("Registering device client...")
	dc, err := gosso.RegisterClient(ctx, oidcClient, "login")
	if err != nil {
		return err
	}
	fmt.Println("done")

	// AUthenticate (incognito browser) with spinner
	token, err := authenticateWithSpinner(ctx, oidcClient, dc, startURL)
	if err != nil {
		return err
	}
	token.Region = region

	// List accounts with spinner
	ssoClient := sso.NewFromConfig(cfg)
	fmt.Print("Fetching accounts...")
	accounts, err := gosso.ListAccounts(ctx, ssoClient, token.AccessToken)
	if err != nil {
		return err
	}
	fmt.Println("done")

	sort.Slice(accounts, func(i, j int) bool { return accounts[i].Name < accounts[j].Name })
	accountLabels := make([]string, len(accounts))
	for i, a := range accounts {
		accountLabels[i] = a.String()
	}
	accountIdx := promptSelectIdx("Select account", accountLabels)
	selected := accounts[accountIdx]

	// List roles
	roles, err := gosso.ListRoles(ctx, ssoClient, token.AccessToken, selected.ID)
	if err != nil {
		return err
	}
	sort.Strings(roles)
	role := promptSelect("Select role", roles)

	// Get credentials
	fmt.Print("Fetching credentials...")
	creds, err := gosso.GetRoleCredentials(ctx, ssoClient, token.AccessToken, selected.ID, role)
	if err != nil {
		return err
	}
	fmt.Println("done")

	// Prompt for profile name
	profileName := promptText("Profile name")

	// Write credentials and cache token
	if err := gosso.WriteCredentials(profileName, creds, region); err != nil {
		return err
	}
	gosso.SaveToken(profileName, token)

	fmt.Printf("\n✓ Credentials written to ~/.aws/credentials [%s]\n", profileName)
	fmt.Println(
		"\n⚠ Close all InPrivate/Incognito browser windows before logging into another environment.",
	)
	fmt.Print("Press Enter after closing the browser...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	return nil
}

func authenticateWithSpinner(
	ctx context.Context,
	client *ssooidc.Client,
	dc *gosso.DeviceClient,
	startURL string,
) (*gosso.AccessToken, error) {
	done := make(chan struct{})
	var token *gosso.AccessToken
	var authErr error

	go func() {
		token, authErr = gosso.Authenticate(ctx, client, dc, startURL)
		close(done)
	}()

	// Spinner while waiting for browser auth
	i := 0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	fmt.Print("Waiting for browser authorization...")
	for {
		select {
		case <-done:
			fmt.Print("\r\033[K") // clear spinner line
			if authErr != nil {
				return nil, authErr
			}
			fmt.Println("✓ Authenticated")
			return token, nil
		case <-ticker.C:
			fmt.Printf(
				"\r%s Waiting for browser authorization...",
				spinnerFrames[i%len(spinnerFrames)],
			)
			i++
		}
	}
}

func promptText(label string) string {
	p := promptui.Prompt{Label: label}
	val, err := p.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "prompt failed: %v\n", err)
		os.Exit(1)
	}
	return val
}

func promptSelect(label string, items []string) string {
	s := promptui.Select{Label: label, Items: items, Size: 15}
	_, val, err := s.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "prompt failed: %v\n", err)
		os.Exit(1)
	}
	return val
}

func promptSelectIdx(label string, items []string) int {
	s := promptui.Select{Label: label, Items: items, Size: 15}
	idx, _, err := s.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "prompt failed: %v\n", err)
		os.Exit(1)
	}
	return idx
}
