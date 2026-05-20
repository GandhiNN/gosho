package cmd

import (
	"fmt"
	"time"

	gosso "github.com/gandhinn/gosho/sso"
)

func Status() error {
	profiles := gosso.ListProfiles()
	if len(profiles) == 0 {
		fmt.Println("No cached profiles. Run 'gosho' to login.")
		return nil
	}

	fmt.Printf("%-20s %-10s %s\n", "PROFILE", "STATUS", "EXPIRES")
	fmt.Printf("%-20s %-10s %s\n", "-------", "------", "-------")

	for _, p := range profiles {
		token, err := gosso.LoadCachedToken(p)
		if err != nil {
			fmt.Printf("%-20s %-10s %s\n", p, "ERROR", "cannot read cache")
			continue
		}
		if token.IsExpired() {
			fmt.Printf(
				"%-20s %-10s %s\n",
				p,
				"EXPIRED",
				token.ExpiresAt.Local().Format(time.DateTime),
			)
		} else {
			remaining := time.Until(token.ExpiresAt).Truncate(time.Minute)
			fmt.Printf("%-20s %-10s %s (%s remaining)\n", p, "VALID", token.ExpiresAt.Local().Format(time.DateTime), remaining)
		}
	}
	return nil
}
