package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/gandhinn/gosho/config"
	gosso "github.com/gandhinn/gosho/sso"
	"github.com/manifoldco/promptui"
)

// Import writes manually-supplied AWS credentials (e.g. copied from an SSO
// portal opened inside an RDP session) into ~/.aws/credentials.
//
// It supports two input methods:
//   - separate prompts for each value (default)
//   - a single pasted credential block via the --paste flag
//
// Imported credentials are static STS keys with no refresh token, so they are
// intentionally NOT added to the token cache/index and will not appear in
// `gosho status`.
func Import(profileArg string, paste bool) error {
	cfg := config.Load()

	profile := profileArg
	if profile == "" {
		profile = promptText("Profile name")
	}

	var creds *gosso.RoleCredentials
	var err error
	if paste {
		creds, err = readPastedCredentials(os.Stdin)
	} else {
		creds, err = readPromptedCredentials()
	}
	if err != nil {
		return err
	}

	region := cfg.Region
	if region == "" {
		region = promptSelect("Region", gosso.Regions)
	} else {
		fmt.Printf("Using region: %s\n", region)
	}

	if err := gosso.WriteCredentials(profile, creds, region); err != nil {
		return err
	}

	fmt.Printf("\n Credentials written to ~/.aws/credentials [%s]\n", profile)
	fmt.Println(" (shown as IMPORTED in 'gosho status'; expiry is not tracked)")

	return nil
}

func readPromptedCredentials() (*gosso.RoleCredentials, error) {
	accessKeyID := strings.TrimSpace(promptText("AWS Access Key ID"))
	secretAccessKey := strings.TrimSpace(promptMasked("AWS Secret Access Key"))
	sessionToken := strings.TrimSpace(promptMasked("AWS Session Token"))

	return validateCredentials(&gosso.RoleCredentials{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		SessionToken:    sessionToken,
	})
}

func readPastedCredentials(r io.Reader) (*gosso.RoleCredentials, error) {
	fmt.Println("Paste the credentials block below, then press Ctrl+Z + Enter (Windows) or Ctrl+D (Unix):")
	data, err := io.ReadAll(bufio.NewReader(r))
	if err != nil {
		return nil, fmt.Errorf("read paste input: %w", err)
	}
	creds := ParseCredentialBlock(string(data))
	return validateCredentials(creds)
}

// ParseCredentialBlock extracts AWS credential values from a pasted block.
// It recognizes the common formats emitted by the AWS console and CLI:
//
//	export AWS_ACCESS_KEY_ID="..."      (bash/zsh)
//	set AWS_ACCESS_KEY_ID=...            (Windows cmd)
//	$Env:AWS_ACCESS_KEY_ID="..."         (PowerShell)
//	aws_access_key_id = ...              (credentials ini)
//
// Matching is case-insensitive on the key names and tolerant of surrounding
// quotes and whitespace.
func ParseCredentialBlock(block string) *gosso.RoleCredentials {
	return &gosso.RoleCredentials{
		AccessKeyID:     extractValue(block, "AWS_ACCESS_KEY_ID"),
		SecretAccessKey: extractValue(block, "AWS_SECRET_ACCESS_KEY"),
		SessionToken:    extractValue(block, "AWS_SESSION_TOKEN"),
	}
}

func extractValue(block, key string) string {
	// Matches an optional prefix (export / set / $Env:), the key, a separator
	// (= or :), and an optionally-quoted value up to end of line.
	pattern := fmt.Sprintf(
		`(?im)^\s*(?:export\s+|set\s+|\$env:)?%s\s*[:=]\s*["']?([^"'\r\n]+)["']?\s*$`,
		regexp.QuoteMeta(key),
	)
	re := regexp.MustCompile(pattern)
	if m := re.FindStringSubmatch(block); m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func validateCredentials(creds *gosso.RoleCredentials) (*gosso.RoleCredentials, error) {
	var missing []string
	if creds.AccessKeyID == "" {
		missing = append(missing, "AWS Access Key ID")
	}
	if creds.SecretAccessKey == "" {
		missing = append(missing, "AWS Secret Access Key")
	}
	if creds.SessionToken == "" {
		missing = append(missing, "AWS Session Token")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing credential value(s): %s", strings.Join(missing, ", "))
	}
	return creds, nil
}

func promptMasked(label string) string {
	p := promptui.Prompt{Label: label, Mask: '*'}
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
