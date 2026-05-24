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

func Login(profileArg string) error {
	ctx := context.Background()
	cfg := config.Load()

	// Determine start URL and region
	startURL := cfg.StartURL
	region := cfg.Region
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
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return fmt.Errorf("load AWS config: %w", err)
	}

	// Register OIDC client
	oidcClient := ssooidc.NewFromConfig(awsCfg)

	// Try cached/refreshed token befor opening browser
	var token *gosso.AccessToken
	if profileArg != "" {
		if cached, err := gosso.LoadCachedToken(profileArg); err == nil {
			if !cached.IsExpired() {
				fmt.Println("Using cached token (still valid)")
				token = cached
			} else if cached.RefreshToken != "" {
				fmt.Print("Refreshing token...")
				refreshed, err := gosso.RefreshToken(ctx, oidcClient, cached)
				if err == nil {
					fmt.Println("done")
					token = refreshed
					gosso.SaveToken(profileArg, token)
				} else {
					fmt.Println("failed, opening browser")
				}
			}
		}
	}

	// Fresh auth if no valid token
	browserOpened := false
	if token == nil {
		fmt.Print("Registering device client...")
		dc, err := gosso.RegisterClient(ctx, oidcClient, "login")
		if err != nil {
			return err
		}
		fmt.Println("done")

		token, err = authenticateWithSpinner(ctx, oidcClient, dc, startURL)
		if err != nil {
			return err
		}
		browserOpened = true
	}
	token.Region = region

	ssoClient := sso.NewFromConfig(awsCfg)

	// If profile preset exists, use it directly
	if profileArg != "" {
		if preset, ok := cfg.Profiles[profileArg]; ok {
			return loginWithPreset(ctx, ssoClient, token, profileArg, preset, region, browserOpened)
		}
		fmt.Printf("Profile %q not found in config, falling back to interactive.\n\n", profileArg)
	}

	// Interactive: list accounts
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

	// Save preset to config for future use
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]config.Profile)
	}
	cfg.Profiles[profileName] = config.Profile{
		AccountID: selected.ID,
		Role:      role,
	}
	config.Save(cfg)

	return writeAndFinish(profileName, creds, region, token, browserOpened)
}

func loginWithPreset(
	ctx context.Context,
	client *sso.Client,
	token *gosso.AccessToken,
	profileName string,
	preset config.Profile,
	region string,
	browserOpened bool,
) error {
	fmt.Printf("Using preset: account=%s, role=%s\n", preset.AccountID, preset.Role)
	fmt.Print("Fetching credentials...")
	creds, err := gosso.GetRoleCredentials(
		ctx,
		client,
		token.AccessToken,
		preset.AccountID,
		preset.Role,
	)
	if err != nil {
		return err
	}
	fmt.Println("done")
	return writeAndFinish(profileName, creds, region, token, browserOpened)
}

func writeAndFinish(
	profileName string,
	creds *gosso.RoleCredentials,
	region string,
	token *gosso.AccessToken,
	browserOpened bool,
) error {
	if err := gosso.WriteCredentials(profileName, creds, region); err != nil {
		return err
	}
	gosso.SaveToken(profileName, token)

	fmt.Printf("\n✓ Credentials written to ~/.aws/credentials [%s]\n", profileName)
	if browserOpened {
		fmt.Println(
			"\n⚠ Close all InPrivate/Incognito browser windows before logging into another environment.",
		)
		fmt.Print("Press Enter after closing the browser...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
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

func LoginAll() error {
	cfg := config.Load()
	if len(cfg.Profiles) == 0 {
		fmt.Println("No saved profiles. Run 'gosho login' first to create one.")
		return nil
	}

	profiles := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		profiles = append(profiles, name)
	}

	for i, name := range profiles {
		fmt.Printf("\n━━━ [%d/%d] Logging in: %s ━━━\n\n", i+1, len(profiles), name)
		if err := Login(name); err != nil {
			fmt.Printf("⚠ Failed to login %s: %v\n", name, err)
		}
	}
	return nil
}
