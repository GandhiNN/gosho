package main

import (
	"fmt"
	"os"

	"github.com/gandhinn/gosho/cmd"
)

func main() {
	if err := cmd.Login(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
