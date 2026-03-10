// Package registry manages the project registry at ~/.modula/configs.json.
// The registry maps project names to environments, each containing an absolute
// modula.config.json path. All credentials, URLs, and settings stay in each project's
// own config — not in the registry.
package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Project holds the environments for a single registered project.
type Project struct {
	Envs       map[string]string `json:"envs"`
	DefaultEnv string            `json:"default_env"`
}

// Registry maps project names to their environments and config paths.
type Registry struct {
	Projects map[string]*Project `json:"projects"`
	Default  string              `json:"default"`
}

// Path returns the absolute path to the registry file (~/.modula/configs.json).
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".modula", "configs.json"), nil
}

// Load reads the registry from disk. Returns an empty Registry if the file
// does not exist. Returns an error only if the file exists but is unreadable
// or contains invalid JSON.
func Load() (*Registry, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{Projects: make(map[string]*Project)}, nil
		}
		return nil, fmt.Errorf("reading registry %s: %w", p, err)
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parsing registry %s: %w", p, err)
	}
	if reg.Projects == nil {
		reg.Projects = make(map[string]*Project)
	}
	return &reg, nil
}

// Save writes the registry to disk, creating ~/.modula/ if needed.
func (r *Registry) Save() error {
	p, err := Path()
	if err != nil {
		return err
	}

	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling registry: %w", err)
	}

	if err := os.WriteFile(p, data, 0600); err != nil {
		return fmt.Errorf("writing registry %s: %w", p, err)
	}
	return nil
}

// Resolve returns the absolute config path for the given project and environment.
// If name is empty, it uses the default project. If env is empty, it uses the
// project's default environment. Returns an error if the project or environment
// is not found.
func (r *Registry) Resolve(name, env string) (string, error) {
	if name == "" {
		if r.Default == "" {
			return "", fmt.Errorf("no project name specified and no default set")
		}
		name = r.Default
	}

	proj, ok := r.Projects[name]
	if !ok {
		return "", fmt.Errorf("project %q not found in registry", name)
	}

	if env == "" {
		if proj.DefaultEnv == "" {
			return "", fmt.Errorf("project %q has no default environment set", name)
		}
		env = proj.DefaultEnv
	}

	p, ok := proj.Envs[env]
	if !ok {
		return "", fmt.Errorf("environment %q not found in project %q", env, name)
	}
	return p, nil
}

// Set creates or updates an environment for a project. The configPath is resolved
// to an absolute path before storing. If the project does not exist, it is created
// and the new env becomes the default. If the environment already exists, its path
// is overwritten. Returns an error if the config file does not exist.
func (r *Registry) Set(name, env, configPath string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if env == "" {
		return fmt.Errorf("environment name cannot be empty")
	}

	abs, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("resolving path %q: %w", configPath, err)
	}

	if _, err := os.Stat(abs); err != nil {
		return fmt.Errorf("config file not found: %s", abs)
	}

	proj, ok := r.Projects[name]
	if !ok {
		proj = &Project{Envs: make(map[string]string)}
		r.Projects[name] = proj
	}

	proj.Envs[env] = abs

	// First env set becomes the default
	if proj.DefaultEnv == "" {
		proj.DefaultEnv = env
	}

	return r.Save()
}

// Remove removes an entire project (all environments) from the registry.
// If the removed project was the default, the default is cleared.
func (r *Registry) Remove(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	if _, ok := r.Projects[name]; !ok {
		return fmt.Errorf("project %q not found in registry", name)
	}

	delete(r.Projects, name)
	if r.Default == name {
		r.Default = ""
	}
	return r.Save()
}

// RemoveEnv removes a single environment from a project. If the removed env was
// the project's default, the default is cleared. If no environments remain, the
// entire project is removed.
func (r *Registry) RemoveEnv(name, env string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if env == "" {
		return fmt.Errorf("environment name cannot be empty")
	}

	proj, ok := r.Projects[name]
	if !ok {
		return fmt.Errorf("project %q not found in registry", name)
	}

	if _, ok := proj.Envs[env]; !ok {
		return fmt.Errorf("environment %q not found in project %q", env, name)
	}

	delete(proj.Envs, env)

	if proj.DefaultEnv == env {
		proj.DefaultEnv = ""
	}

	// Remove the project entirely if no envs remain
	if len(proj.Envs) == 0 {
		delete(r.Projects, name)
		if r.Default == name {
			r.Default = ""
		}
	}

	return r.Save()
}

// SetDefault sets the default project name. The project must already exist
// in the registry.
func (r *Registry) SetDefault(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	if _, ok := r.Projects[name]; !ok {
		return fmt.Errorf("project %q not found in registry — set it first", name)
	}

	r.Default = name
	return r.Save()
}

// SetDefaultEnv sets the default environment for a project. The environment
// must already exist in the project.
func (r *Registry) SetDefaultEnv(name, env string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if env == "" {
		return fmt.Errorf("environment name cannot be empty")
	}

	proj, ok := r.Projects[name]
	if !ok {
		return fmt.Errorf("project %q not found in registry", name)
	}

	if _, ok := proj.Envs[env]; !ok {
		return fmt.Errorf("environment %q not found in project %q", env, name)
	}

	proj.DefaultEnv = env
	return r.Save()
}

// EnvNames returns the sorted environment names for a project.
func (r *Registry) EnvNames(name string) []string {
	proj, ok := r.Projects[name]
	if !ok {
		return nil
	}
	names := make([]string, 0, len(proj.Envs))
	for env := range proj.Envs {
		names = append(names, env)
	}
	sort.Strings(names)
	return names
}
