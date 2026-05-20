package main

import (
	"fmt"
	"os"

	"github.com/gandhinn/gosho/cmd"
)

const usage = `gosho - AWS SSO login with fresh browser sessions

Usage:
	gosho			Login to AWS SSO (interactive)
	gosho init 		Configure default start URL and region
	gosho status	Show cached profile status
	gosho --help 	Show this help message
`

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			run(cmd.Init)
		case "status":
			run(cmd.Status)
		case "--help", "-h", "help":
			fmt.Print(usage)
		default:
			fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
			fmt.Print(usage)
			os.Exit(1)
		}
		return
	}
	run(cmd.Login)
}

func run(fn func() error) {
	if err := fn(); err != nil {
		if err.Error() == "^C" {
			fmt.Println()
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
