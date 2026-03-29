package mcp

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/registry"
	modula "github.com/hegner123/modulacms/sdks/go"
)

// ConnectionManager holds the project registry and manages the active SDK
// client. MCP tools call SwitchProject to change the target CMS instance at
// runtime without restarting the server.
type ConnectionManager struct {
	mu       sync.RWMutex
	registry *registry.Registry
	client   *modula.Client // current active client (nil until connected)
	project  string         // active project name
	env      string         // active environment name
	url      string         // active base URL
}

// NewConnectionManager loads the registry from ~/.modula/configs.json and
// returns a manager with no active connection.
func NewConnectionManager() (*ConnectionManager, error) {
	reg, err := registry.Load()
	if err != nil {
		return nil, fmt.Errorf("loading project registry: %w", err)
	}
	return &ConnectionManager{registry: reg}, nil
}

// SwitchProject reads the config for the given project/environment from the
// registry, extracts the port and MCP API key, and creates a new SDK client.
func (cm *ConnectionManager) SwitchProject(project, env string) error {
	resolved, err := cm.registry.Resolve(project, env)
	if err != nil {
		return fmt.Errorf("resolve project %q env %q: %w", project, env, err)
	}

	cfg, loadErr := loadConfigFromPaths(resolved.Base, resolved.Overlay)
	if loadErr != nil {
		return fmt.Errorf("load config for %q: %w", project, loadErr)
	}

	apiKey := cfg.MCP_API_Key
	if apiKey == "" {
		return fmt.Errorf("project %q has no mcp_api_key configured", project)
	}

	baseURL := buildBaseURL(cfg)

	client, err := modula.NewClient(modula.ClientConfig{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		HTTPClient: mcpHTTPClient(baseURL),
	})
	if err != nil {
		return fmt.Errorf("create SDK client for %q: %w", project, err)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.client = client
	cm.project = project
	cm.env = env
	cm.url = baseURL
	return nil
}

// Client returns the active SDK client. Returns nil if no connection is active.
func (cm *ConnectionManager) Client() *modula.Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.client
}

// ActiveConnection returns the current project, environment, and URL.
func (cm *ConnectionManager) ActiveConnection() (project, env, url string) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.project, cm.env, cm.url
}

// ListProjects returns a JSON-serializable summary of all registered projects.
func (cm *ConnectionManager) ListProjects() (json.RawMessage, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	type envInfo struct {
		Name      string `json:"name"`
		IsDefault bool   `json:"is_default"`
	}
	type projectInfo struct {
		Name         string    `json:"name"`
		IsDefault    bool      `json:"is_default"`
		IsActive     bool      `json:"is_active"`
		Environments []envInfo `json:"environments"`
	}

	var projects []projectInfo
	for name, proj := range cm.registry.Projects {
		pi := projectInfo{
			Name:      name,
			IsDefault: name == cm.registry.Default,
			IsActive:  name == cm.project,
		}
		for envName := range proj.Envs {
			pi.Environments = append(pi.Environments, envInfo{
				Name:      envName,
				IsDefault: envName == proj.DefaultEnv,
			})
		}
		projects = append(projects, pi)
	}

	return json.Marshal(projects)
}

// loadConfigFromPaths reads a base config and optionally merges an overlay.
func loadConfigFromPaths(basePath, overlayPath string) (*config.Config, error) {
	if overlayPath != "" {
		provider := config.NewLayeredFileProvider(basePath, overlayPath)
		return provider.Get()
	}
	provider := config.NewFileProvider(basePath)
	return provider.Get()
}

// buildBaseURL derives the MCP connection URL from the config.
// Priority: mcp_url (explicit) > localhost:port (fallback).
func buildBaseURL(cfg *config.Config) string {
	if cfg.MCP_URL != "" {
		return strings.TrimRight(cfg.MCP_URL, "/")
	}
	host := cfg.Environment.HTTPHost()
	if host == "" {
		host = "localhost"
	}
	port := cfg.Port
	if port == "" {
		port = ":8080"
	}
	if port == ":80" {
		return "http://" + host
	}
	return "http://" + host + port
}

// mcpHTTPClient returns an http.Client that injects X-Modula-MCP: true on
// every request. For https://localhost URLs, TLS certificate verification is
// disabled to support self-signed development certificates.
func mcpHTTPClient(baseURL string) *http.Client {
	var base http.RoundTripper = http.DefaultTransport
	if strings.HasPrefix(baseURL, "https://localhost") {
		base = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: &mcpTransport{base: base},
	}
}

// mcpTransport injects X-Modula-MCP: true into every outgoing request.
type mcpTransport struct {
	base http.RoundTripper
}

func (t *mcpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("X-Modula-MCP", "true")
	return t.base.RoundTrip(req)
}

// ReloadRegistry re-reads ~/.modula/configs.json from disk.
func (cm *ConnectionManager) ReloadRegistry() error {
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("reloading registry: %w", err)
	}
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.registry = reg
	return nil
}

// ConnectFromCwd attempts to find and connect to a project that matches the
// current working directory. Returns false (without error) if no match is found.
func (cm *ConnectionManager) ConnectFromCwd() (bool, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return false, nil
	}

	cm.mu.RLock()
	name, proj := cm.registry.FindByDir(cwd)
	cm.mu.RUnlock()

	if proj == nil {
		return false, nil
	}

	env := proj.DefaultEnv
	if err := cm.SwitchProject(name, env); err != nil {
		return false, err
	}
	return true, nil
}
