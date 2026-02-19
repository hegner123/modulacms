package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

const version = "0.1.0"

func main() {
	url := os.Getenv("MODULACMS_URL")
	apiKey := os.Getenv("MODULACMS_API_KEY")

	if url == "" || apiKey == "" {
		fmt.Fprintln(os.Stderr, "Usage: MODULACMS_URL and MODULACMS_API_KEY environment variables are required")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "  MODULACMS_URL     Base URL of the ModulaCMS server (e.g. http://localhost:8080)")
		fmt.Fprintln(os.Stderr, "  MODULACMS_API_KEY Bearer token for API authentication")
		os.Exit(1)
	}

	client, err := modulacms.NewClient(modulacms.ClientConfig{
		BaseURL: url,
		APIKey:  apiKey,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create SDK client: %v\n", err)
		os.Exit(1)
	}

	srv := server.NewMCPServer("modulacms", version)

	registerContentTools(srv, client)
	registerSchemaTools(srv, client)
	registerMediaTools(srv, client)
	registerRouteTools(srv, client)
	registerUserTools(srv, client)
	registerRBACTools(srv, client)
	registerConfigTools(srv, client)
	registerImportTools(srv, client)

	if err := server.ServeStdio(srv); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
