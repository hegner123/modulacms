// Package main is the entry point for the ModulaCMS application.
// It initializes the CLI and starts the HTTP/HTTPS/SSH servers based on
// command-line arguments and configuration.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
