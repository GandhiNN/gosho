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
	colorReset  = "\033[0m"
)

func Status() error {
	profiles := gosso.ListProfiles()
	if len(profiles) == 0 {
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
	return nil
}
