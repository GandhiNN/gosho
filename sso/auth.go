package sso

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
)

func RegisterClient(
	ctx context.Context,
	client *ssooidc.Client,
	sessionName string,
) (*DeviceClient, error) {
	resp, err := client.RegisterClient(ctx, &ssooidc.RegisterClientInput{
		ClientName: aws.String(fmt.Sprintf("%s-%s", ClientName, sessionName)),
		ClientType: aws.String("public"),
		Scopes:     []string{DefaultScope},
	})
	if err != nil {
		return nil, fmt.Errorf("register client: %w", err)
	}
	return &DeviceClient{
		ClientID:     aws.ToString(resp.ClientId),
		ClientSecret: aws.ToString(resp.ClientSecret),
		ExpiresAt:    time.Unix(resp.ClientSecretExpiresAt, 0).UTC(),
	}, nil
}

func Authenticate(
	ctx context.Context,
	client *ssooidc.Client,
	dc *DeviceClient,
	startURL string,
) (*AccessToken, error) {
	authResp, err := client.StartDeviceAuthorization(ctx, &ssooidc.StartDeviceAuthorizationInput{
		ClientId:     &dc.ClientID,
		ClientSecret: &dc.ClientSecret,
		StartUrl:     &startURL,
	})
	if err != nil {
		return nil, fmt.Errorf("start device auth: %w", err)
	}

	verifyURL := aws.ToString(authResp.VerificationUriComplete)
	openIncognito(verifyURL)
	fmt.Printf("\nVerify code: \033[36;1m%s\033[0m\n", aws.ToString(authResp.UserCode))
	fmt.Printf("If browser did not open: %s\n\n", verifyURL)

	interval := time.Duration(authResp.Interval) * time.Second
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("authentication cancelled")
		default:
		}
		tokenResp, err := client.CreateToken(ctx, &ssooidc.CreateTokenInput{
			ClientId:     &dc.ClientID,
			ClientSecret: &dc.ClientSecret,
			GrantType:    aws.String(DeviceGrantType),
			DeviceCode:   authResp.DeviceCode,
		})
		if err != nil {
			if isAuthPending(err) {
				time.Sleep(interval)
				continue
			}
			if isAccessDenied(err) {
				return nil, fmt.Errorf("access request rejected")
			}
			return nil, fmt.Errorf("create token: %w", err)
		}

		expiresAt := time.Now().UTC().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		return &AccessToken{
			StartURL:     startURL,
			Region:       "",
			AccessToken:  aws.ToString(tokenResp.AccessToken),
			ExpiresAt:    expiresAt,
			RefreshToken: aws.ToString(tokenResp.RefreshToken),
			Client:       *dc,
		}, nil
	}
}

func RefreshToken(
	ctx context.Context,
	client *ssooidc.Client,
	token *AccessToken,
) (*AccessToken, error) {
	resp, err := client.CreateToken(ctx, &ssooidc.CreateTokenInput{
		ClientId:     &token.Client.ClientID,
		ClientSecret: &token.Client.ClientSecret,
		GrantType:    aws.String(RefreshGrantType),
		RefreshToken: &token.RefreshToken,
	})
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}
	expiresAt := time.Now().UTC().Add(time.Duration(resp.ExpiresIn) * time.Second)
	return &AccessToken{
		StartURL:     token.StartURL,
		Region:       token.Region,
		AccessToken:  aws.ToString(resp.AccessToken),
		ExpiresAt:    expiresAt,
		RefreshToken: aws.ToString(resp.RefreshToken),
		Client:       token.Client,
	}, nil
}

func isAuthPending(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "AuthorizationPendingException") ||
		strings.Contains(err.Error(), "authorization_pending"))
}

func isAccessDenied(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "AccessDeniedException") ||
		strings.Contains(err.Error(), "access_denied"))
}

func openIncognito(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		// Try common browsers in incognito/private mode
		if isWSL() {
			cmd = exec.Command("cmd.exe", "/c", "start", "msedge", "--inprivate", url)
		} else if path, err := exec.LookPath("google-chrome"); err == nil {
			cmd = exec.Command(path, "--incognito", url)
		} else if path, err := exec.LookPath("microsoft-edge"); err == nil {
			cmd = exec.Command(path, "--inprivate", url)
		} else if path, err := exec.LookPath("firefox"); err == nil {
			cmd = exec.Command(path, "--private-window", url)
		} else {
			cmd = exec.Command("xdg-open", url)
		}
	case "darwin":
		cmd = exec.Command("open", "-na", "Google Chrome", "--args", "--incognito", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "msedge", "--inprivate", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}

func isWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), "microsoft")
}
