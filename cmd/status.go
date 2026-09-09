package cmd

import (
	"fmt"
	"time"

	gosso "github.com/gandhinn/gosho/sso"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorGray   = "\033[90m"
	colorReset  = "\033[0m"
)

func Status() error {
	profiles := gosso.ListProfiles()
	imported := importedProfiles(profiles)

	if len(profiles) == 0 && len(imported) == 0 {
		fmt.Println("No cached profiles. Run 'gosho login' to login.")
		return nil
	}

	fmt.Printf("%-20s %-10s %s\n", "PROFILE", "STATUS", "EXPIRES")
	fmt.Printf("%-20s %-10s %s\n", "-------", "------", "-------")

	for _, p := range profiles {
		token, err := gosso.LoadCachedToken(p)
		if err != nil {
			fmt.Printf(
				"%-20s %s%-10s%s %s\n",
				p,
				colorRed,
				"ERROR",
				colorReset,
				"cannot read cache",
			)
			continue
		}
		if token.IsExpired() {
			fmt.Printf(
				"%-20s %s%-10s%s %s\n",
				p,
				colorRed,
				"EXPIRED",
				colorReset,
				token.ExpiresAt.Local().Format(time.DateTime),
			)
		} else {
			remaining := time.Until(token.ExpiresAt).Truncate(time.Minute)
			color := colorGreen
			if remaining < 30*time.Minute {
				color = colorYellow
			}
			fmt.Printf("%-20s %s%-10s%s %s (%s remaining)\n", p, color, "VALID", colorReset, token.ExpiresAt.Local().Format(time.DateTime), remaining)
		}
	}

	for _, p := range imported {
		fmt.Printf(
			"%-20s %s%-10s%s %s\n",
			p,
			colorGray,
			"IMPORTED",
			colorReset,
			"expiry unknown",
		)
	}
	return nil
}

// importedProfiles returns profiles present in the AWS credentials file that
// are not tracked in the token cache index. These were written via
// `gosho import` (or externally) and have no expiry information available.
func importedProfiles(tracked []string) []string {
	trackedSet := make(map[string]struct{}, len(tracked))
	for _, p := range tracked {
		trackedSet[p] = struct{}{}
	}
	var imported []string
	for _, p := range gosso.ListCredentialProfiles() {
		if _, ok := trackedSet[p]; !ok {
			imported = append(imported, p)
		}
	}
	return imported
}
