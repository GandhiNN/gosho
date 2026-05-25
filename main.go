package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gandhinn/gosho/cmd"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "login":
			profile := ""
			if len(os.Args) > 2 {
				profile = os.Args[2]
			}
			if profile == "all" {
				runE(cmd.LoginAll())
			} else if strings.HasPrefix(profile, "-") {
				printUsage()
			} else {
				runE(cmd.Login(profile))
			}
		case "logout":
			profile := ""
			if len(os.Args) > 2 {
				profile = os.Args[2]
			}
			runE(cmd.Logout(profile))
		case "init":
			runE(cmd.Init())
		case "status":
			runE(cmd.Status())
		case "--help", "-h", "help":
			printUsage()
		case "--version", "-v", "version":
			fmt.Printf("gosho %s (commit: %s, built: %s)\n", version, commit, date)
		default:
			fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
			printUsage()
			os.Exit(1)
		}
		return
	}
	printUsage()
}

func runE(err error) {
	if err != nil {
		if err.Error() == "^C" {
			fmt.Println()
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	lines := []struct{ cmd, desc string }{
		{"gosho init", "Configure default start URL and region"},
		{"gosho login [profile]", "Login to AWS SSO (use preset if profile exists)"},
		{"gosho login all", "Login to all saved profiles"},
		{"gosho logout [profile]", "Clear cached token and credentials for a profile"},
		{"gosho logout all", "Clear all cached token and credentials"},
		{"gosho status", "Show cached profile status"},
		{"gosho help", "Show this help message"},
	}
	fmt.Println("gosho - AWS SSO login with fresh browser sessions")
	fmt.Println("\nUsage:")
	for _, l := range lines {
		fmt.Printf(" %-24s %s\n", l.cmd, l.desc)
	}
}
