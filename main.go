package main

import (
	"fmt"
	"os"

	"github.com/gandhinn/gosho/cmd"
)

const usage = `gosho - AWS SSO login with fresh browser sessions

Usage:
	gosho [profile]		Login to AWS SSO (use preset if profile defined in config)
	gosho init 			Configure default start URL and region
	gosho status		Show cached profile status
	gosho --help 		Show this help message
`

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			runE(cmd.Init())
		case "status":
			runE(cmd.Status())
		case "--help", "-h", "help":
			fmt.Print(usage)
		default:
			// Treat unknown arg as profile name
			runE(cmd.Login(os.Args[1]))
		}
		return
	}
	runE(cmd.Login(""))
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
