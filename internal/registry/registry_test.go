package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// setTestHome points HOME at a temp directory so Load/Save/Set/Remove/etc.
// operate on an isolated ~/.modula/configs.json instead of the real one.
func setTestHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	return tmp
}

// writeRegistry is a helper that writes a Registry to the test home.
func writeRegistry(t *testing.T, reg *Registry) {
	t.Helper()
	p, err := Path()
	if err != nil {
		t.Fatalf("Path(): %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0750); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(p, data, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// createConfigFile creates a dummy config.json at the given path.
func createConfigFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(`{"db_driver":"sqlite"}`), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func TestPath(t *testing.T) {
	home := setTestHome(t)
	p, err := Path()
	if err != nil {
		t.Fatalf("Path() error: %v", err)
	}
	want := filepath.Join(home, ".modula", "configs.json")
	if p != want {
		t.Errorf("Path() = %q, want %q", p, want)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	setTestHome(t)

	reg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if reg == nil {
		t.Fatal("Load() returned nil")
	}
	if len(reg.Projects) != 0 {
		t.Errorf("expected empty projects, got %d", len(reg.Projects))
	}
	if reg.Default != "" {
		t.Errorf("expected empty default, got %q", reg.Default)
	}
}

func TestLoad_ValidFile(t *testing.T) {
	setTestHome(t)

	writeRegistry(t, &Registry{
		Projects: map[string]*Project{
			"mysite": {
				Envs:       map[string]string{"local": "/tmp/config.json"},
				DefaultEnv: "local",
			},
		},
		Default: "mysite",
	})

	reg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if reg.Default != "mysite" {
		t.Errorf("Default = %q, want %q", reg.Default, "mysite")
	}
	proj, ok := reg.Projects["mysite"]
	if !ok {
		t.Fatal("project 'mysite' missing")
	}
	if proj.DefaultEnv != "local" {
		t.Errorf("DefaultEnv = %q, want %q", proj.DefaultEnv, "local")
	}
	if proj.Envs["local"] != "/tmp/config.json" {
		t.Errorf("Envs[local] = %q, want %q", proj.Envs["local"], "/tmp/config.json")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	setTestHome(t)

	p, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(`{not valid json`), 0600); err != nil {
		t.Fatal(err)
	}

	_, err = Load()
	if err == nil {
		t.Fatal("Load() expected error for invalid JSON, got nil")
	}
}

func TestLoad_NilProjectsField(t *testing.T) {
	setTestHome(t)

	p, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0750); err != nil {
		t.Fatal(err)
	}
	// JSON with no "projects" key — unmarshals to nil map.
	if err := os.WriteFile(p, []byte(`{"default":"x"}`), 0600); err != nil {
		t.Fatal(err)
	}

	reg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if reg.Projects == nil {
		t.Fatal("Projects should be initialized, not nil")
	}
	if len(reg.Projects) != 0 {
		t.Errorf("expected empty projects, got %d", len(reg.Projects))
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	home := setTestHome(t)

	reg := &Registry{Projects: make(map[string]*Project)}
	if err := reg.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	p := filepath.Join(home, ".modula", "configs.json")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("registry file not created: %v", err)
	}
}

func TestSave_RoundTrip(t *testing.T) {
	setTestHome(t)

	original := &Registry{
		Projects: map[string]*Project{
			"blog": {
				Envs:       map[string]string{"dev": "/a/b/config.json", "prod": "/c/d/config.json"},
				DefaultEnv: "dev",
			},
		},
		Default: "blog",
	}
	if err := original.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.Default != "blog" {
		t.Errorf("Default = %q, want %q", loaded.Default, "blog")
	}
	proj := loaded.Projects["blog"]
	if proj == nil {
		t.Fatal("project 'blog' missing after round-trip")
	}
	if proj.Envs["dev"] != "/a/b/config.json" {
		t.Errorf("Envs[dev] = %q", proj.Envs["dev"])
	}
	if proj.Envs["prod"] != "/c/d/config.json" {
		t.Errorf("Envs[prod] = %q", proj.Envs["prod"])
	}
}

func TestResolve(t *testing.T) {
	reg := &Registry{
		Projects: map[string]*Project{
			"site": {
				Envs:       map[string]string{"local": "/x/config.json", "staging": "/y/config.json"},
				DefaultEnv: "local",
			},
		},
		Default: "site",
	}

	t.Run("explicit project and env", func(t *testing.T) {
		p, err := reg.Resolve("site", "staging")
		if err != nil {
			t.Fatalf("Resolve error: %v", err)
		}
		if p != "/y/config.json" {
			t.Errorf("got %q, want %q", p, "/y/config.json")
		}
	})

	t.Run("explicit project default env", func(t *testing.T) {
		p, err := reg.Resolve("site", "")
		if err != nil {
			t.Fatalf("Resolve error: %v", err)
		}
		if p != "/x/config.json" {
			t.Errorf("got %q, want %q", p, "/x/config.json")
		}
	})

	t.Run("default project and env", func(t *testing.T) {
		p, err := reg.Resolve("", "")
		if err != nil {
			t.Fatalf("Resolve error: %v", err)
		}
		if p != "/x/config.json" {
			t.Errorf("got %q, want %q", p, "/x/config.json")
		}
	})

	t.Run("no default project", func(t *testing.T) {
		r := &Registry{Projects: make(map[string]*Project)}
		_, err := r.Resolve("", "")
		if err == nil {
			t.Fatal("expected error for no default project")
		}
	})

	t.Run("project not found", func(t *testing.T) {
		_, err := reg.Resolve("missing", "")
		if err == nil {
			t.Fatal("expected error for missing project")
		}
	})

	t.Run("env not found", func(t *testing.T) {
		_, err := reg.Resolve("site", "production")
		if err == nil {
			t.Fatal("expected error for missing env")
		}
	})

	t.Run("no default env", func(t *testing.T) {
		r := &Registry{
			Projects: map[string]*Project{
				"bare": {Envs: map[string]string{"x": "/z"}, DefaultEnv: ""},
			},
		}
		_, err := r.Resolve("bare", "")
		if err == nil {
			t.Fatal("expected error when project has no default env")
		}
	})
}

func TestSet(t *testing.T) {
	t.Run("new project first env becomes default", func(t *testing.T) {
		home := setTestHome(t)
		configPath := filepath.Join(home, "project", "config.json")
		createConfigFile(t, configPath)

		reg := &Registry{Projects: make(map[string]*Project)}
		if err := reg.Set("mysite", "local", configPath); err != nil {
			t.Fatalf("Set() error: %v", err)
		}

		proj := reg.Projects["mysite"]
		if proj == nil {
			t.Fatal("project not created")
		}
		if proj.DefaultEnv != "local" {
			t.Errorf("DefaultEnv = %q, want %q", proj.DefaultEnv, "local")
		}
		if proj.Envs["local"] != configPath {
			t.Errorf("Envs[local] = %q, want %q", proj.Envs["local"], configPath)
		}
	})

	t.Run("second env does not change default", func(t *testing.T) {
		home := setTestHome(t)
		cfg1 := filepath.Join(home, "a", "config.json")
		cfg2 := filepath.Join(home, "b", "config.json")
		createConfigFile(t, cfg1)
		createConfigFile(t, cfg2)

		reg := &Registry{Projects: make(map[string]*Project)}
		if err := reg.Set("site", "dev", cfg1); err != nil {
			t.Fatal(err)
		}
		if err := reg.Set("site", "staging", cfg2); err != nil {
			t.Fatal(err)
		}

		proj := reg.Projects["site"]
		if proj.DefaultEnv != "dev" {
			t.Errorf("DefaultEnv changed to %q, should remain %q", proj.DefaultEnv, "dev")
		}
		if len(proj.Envs) != 2 {
			t.Errorf("expected 2 envs, got %d", len(proj.Envs))
		}
	})

	t.Run("overwrite existing env path", func(t *testing.T) {
		home := setTestHome(t)
		cfg1 := filepath.Join(home, "old", "config.json")
		cfg2 := filepath.Join(home, "new", "config.json")
		createConfigFile(t, cfg1)
		createConfigFile(t, cfg2)

		reg := &Registry{Projects: make(map[string]*Project)}
		if err := reg.Set("site", "local", cfg1); err != nil {
			t.Fatal(err)
		}
		if err := reg.Set("site", "local", cfg2); err != nil {
			t.Fatal(err)
		}

		if reg.Projects["site"].Envs["local"] != cfg2 {
			t.Errorf("path not overwritten: got %q", reg.Projects["site"].Envs["local"])
		}
	})

	t.Run("config file not found", func(t *testing.T) {
		setTestHome(t)
		reg := &Registry{Projects: make(map[string]*Project)}
		err := reg.Set("site", "local", "/nonexistent/config.json")
		if err == nil {
			t.Fatal("expected error for missing config file")
		}
	})

	t.Run("empty project name", func(t *testing.T) {
		reg := &Registry{Projects: make(map[string]*Project)}
		err := reg.Set("", "local", "/x")
		if err == nil {
			t.Fatal("expected error for empty project name")
		}
	})

	t.Run("empty env name", func(t *testing.T) {
		reg := &Registry{Projects: make(map[string]*Project)}
		err := reg.Set("site", "", "/x")
		if err == nil {
			t.Fatal("expected error for empty env name")
		}
	})

	t.Run("stores absolute path", func(t *testing.T) {
		home := setTestHome(t)
		dir := filepath.Join(home, "project")
		configPath := filepath.Join(dir, "config.json")
		createConfigFile(t, configPath)

		reg := &Registry{Projects: make(map[string]*Project)}
		// Pass relative path (from the dir itself)
		origWd, _ := os.Getwd()
		if err := os.Chdir(dir); err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origWd) //nolint:errcheck // test cleanup

		if err := reg.Set("site", "local", "config.json"); err != nil {
			t.Fatal(err)
		}

		stored := reg.Projects["site"].Envs["local"]
		if !filepath.IsAbs(stored) {
			t.Errorf("stored path is not absolute: %q", stored)
		}
	})

	t.Run("persists to disk", func(t *testing.T) {
		home := setTestHome(t)
		configPath := filepath.Join(home, "proj", "config.json")
		createConfigFile(t, configPath)

		reg := &Registry{Projects: make(map[string]*Project)}
		if err := reg.Set("site", "local", configPath); err != nil {
			t.Fatal(err)
		}

		loaded, err := Load()
		if err != nil {
			t.Fatalf("Load() after Set(): %v", err)
		}
		if _, ok := loaded.Projects["site"]; !ok {
			t.Fatal("project not persisted to disk")
		}
	})
}

func TestRemove(t *testing.T) {
	t.Run("removes project", func(t *testing.T) {
		setTestHome(t)
		reg := &Registry{
			Projects: map[string]*Project{
				"site": {Envs: map[string]string{"local": "/x"}, DefaultEnv: "local"},
			},
		}
		writeRegistry(t, reg)

		if err := reg.Remove("site"); err != nil {
			t.Fatalf("Remove() error: %v", err)
		}
		if _, ok := reg.Projects["site"]; ok {
			t.Error("project still present after Remove")
		}
	})

	t.Run("clears default when removed project was default", func(t *testing.T) {
		setTestHome(t)
		reg := &Registry{
			Projects: map[string]*Project{
				"site": {Envs: map[string]string{"local": "/x"}, DefaultEnv: "local"},
			},
			Default: "site",
		}
		writeRegistry(t, reg)

		if err := reg.Remove("site"); err != nil {
			t.Fatal(err)
		}
		if reg.Default != "" {
			t.Errorf("Default not cleared: %q", reg.Default)
		}
	})

	t.Run("does not clear default for other project", func(t *testing.T) {
		setTestHome(t)
		reg := &Registry{
			Projects: map[string]*Project{
				"a": {Envs: map[string]string{"local": "/x"}, DefaultEnv: "local"},
				"b": {Envs: map[string]string{"local": "/y"}, DefaultEnv: "local"},
			},
			Default: "a",
		}
		writeRegistry(t, reg)

		if err := reg.Remove("b"); err != nil {
			t.Fatal(err)
		}
		if reg.Default != "a" {
			t.Errorf("Default changed to %q, want %q", reg.Default, "a")
		}
	})

	t.Run("not found", func(t *testing.T) {
		setTestHome(t)
		reg := &Registry{Projects: make(map[string]*Project)}
		writeRegistry(t, reg)

		err := reg.Remove("missing")
		if err == nil {
			t.Fatal("expected error for missing project")
		}
	})

	t.Run("empty name", func(t *testing.T) {
		reg := &Registry{Projects: make(map[string]*Project)}
		err := reg.Remove("")
		if err == nil {
			t.Fatal("expected error for empty name")
		}
	})
}

func TestRemoveEnv(t *testing.T) {
	t.Run("removes single env", func(t *testing.T) {
		setTestHome(t)
		reg := &Registry{
			Projects: map[string]*Project{
				"site": {
					Envs:       map[string]string{"dev": "/a", "prod": "/b"},
					DefaultEnv: "dev",
				},
			},
		}
		writeRegistry(t, reg)

		if err := reg.RemoveEnv("site", "prod"); err != nil {
			t.Fatalf("RemoveEnv() error: %v", err)
		}
		if _, ok := reg.Projects["site"].Envs["prod"]; ok {
			t.Error("env 'prod' still present")
		}
		if len(reg.Projects["site"].Envs) != 1 {
			t.Errorf("expected 1 env, got %d", len(reg.Projects["site"].Envs))
		}
	})

	t.Run("clears default env when removed", func(t *testing.T) {
		setTestHome(t)
		reg := &Registry{
			Projects: map[string]*Project{
				"site": {
					Envs:       map[string]string{"dev": "/a", "prod": "/b"},
					DefaultEnv: "dev",
				},
			},
		}
		writeRegistry(t, reg)

		if err := reg.RemoveEnv("site", "dev"); err != nil {
			t.Fatal(err)
		}
		if reg.Projects["site"].DefaultEnv != "" {
			t.Errorf("DefaultEnv not cleared: %q", reg.Projects["site"].DefaultEnv)
		}
	})

	t.Run("last env removes project", func(t *testing.T) {
		setTestHome(t)
		reg := &Registry{
			Projects: map[string]*Project{
				"site": {
					Envs:       map[string]string{"local": "/a"},
					DefaultEnv: "local",
				},
			},
			Default: "site",
		}
		writeRegistry(t, reg)

		if err := reg.RemoveEnv("site", "local"); err != nil {
			t.Fatal(err)
		}
		if _, ok := reg.Projects["site"]; ok {
			t.Error("project should be removed when last env deleted")
		}
		if reg.Default != "" {
			t.Errorf("Default not cleared: %q", reg.Default)
		}
	})

	t.Run("project not found", func(t *testing.T) {
		setTestHome(t)
		reg := &Registry{Projects: make(map[string]*Project)}
		writeRegistry(t, reg)

		err := reg.RemoveEnv("missing", "local")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("env not found", func(t *testing.T) {
		setTestHome(t)
		reg := &Registry{
			Projects: map[string]*Project{
				"site": {Envs: map[string]string{"dev": "/a"}, DefaultEnv: "dev"},
			},
		}
		writeRegistry(t, reg)

		err := reg.RemoveEnv("site", "nonexistent")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("empty names", func(t *testing.T) {
		reg := &Registry{Projects: make(map[string]*Project)}
		if err := reg.RemoveEnv("", "local"); err == nil {
			t.Error("expected error for empty project name")
		}
		if err := reg.RemoveEnv("site", ""); err == nil {
			t.Error("expected error for empty env name")
		}
	})
}

func TestSetDefault(t *testing.T) {
	t.Run("sets default", func(t *testing.T) {
		setTestHome(t)
		reg := &Registry{
			Projects: map[string]*Project{
				"site": {Envs: map[string]string{"local": "/a"}, DefaultEnv: "local"},
			},
		}
		writeRegistry(t, reg)

		if err := reg.SetDefault("site"); err != nil {
			t.Fatalf("SetDefault() error: %v", err)
		}
		if reg.Default != "site" {
			t.Errorf("Default = %q, want %q", reg.Default, "site")
		}
	})

	t.Run("project not found", func(t *testing.T) {
		setTestHome(t)
		reg := &Registry{Projects: make(map[string]*Project)}
		writeRegistry(t, reg)

		err := reg.SetDefault("missing")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("empty name", func(t *testing.T) {
		reg := &Registry{Projects: make(map[string]*Project)}
		err := reg.SetDefault("")
		if err == nil {
			t.Fatal("expected error for empty name")
		}
	})

	t.Run("persists to disk", func(t *testing.T) {
		setTestHome(t)
		reg := &Registry{
			Projects: map[string]*Project{
				"site": {Envs: map[string]string{"local": "/a"}, DefaultEnv: "local"},
			},
		}
		writeRegistry(t, reg)

		if err := reg.SetDefault("site"); err != nil {
			t.Fatal(err)
		}

		loaded, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if loaded.Default != "site" {
			t.Errorf("persisted Default = %q, want %q", loaded.Default, "site")
		}
	})
}

func TestSetDefaultEnv(t *testing.T) {
	t.Run("sets default env", func(t *testing.T) {
		setTestHome(t)
		reg := &Registry{
			Projects: map[string]*Project{
				"site": {
					Envs:       map[string]string{"dev": "/a", "prod": "/b"},
					DefaultEnv: "dev",
				},
			},
		}
		writeRegistry(t, reg)

		if err := reg.SetDefaultEnv("site", "prod"); err != nil {
			t.Fatalf("SetDefaultEnv() error: %v", err)
		}
		if reg.Projects["site"].DefaultEnv != "prod" {
			t.Errorf("DefaultEnv = %q, want %q", reg.Projects["site"].DefaultEnv, "prod")
		}
	})

	t.Run("project not found", func(t *testing.T) {
		setTestHome(t)
		reg := &Registry{Projects: make(map[string]*Project)}
		writeRegistry(t, reg)

		err := reg.SetDefaultEnv("missing", "local")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("env not found", func(t *testing.T) {
		setTestHome(t)
		reg := &Registry{
			Projects: map[string]*Project{
				"site": {Envs: map[string]string{"dev": "/a"}, DefaultEnv: "dev"},
			},
		}
		writeRegistry(t, reg)

		err := reg.SetDefaultEnv("site", "nonexistent")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("empty names", func(t *testing.T) {
		reg := &Registry{Projects: make(map[string]*Project)}
		if err := reg.SetDefaultEnv("", "local"); err == nil {
			t.Error("expected error for empty project name")
		}
		if err := reg.SetDefaultEnv("site", ""); err == nil {
			t.Error("expected error for empty env name")
		}
	})

	t.Run("persists to disk", func(t *testing.T) {
		setTestHome(t)
		reg := &Registry{
			Projects: map[string]*Project{
				"site": {
					Envs:       map[string]string{"dev": "/a", "prod": "/b"},
					DefaultEnv: "dev",
				},
			},
		}
		writeRegistry(t, reg)

		if err := reg.SetDefaultEnv("site", "prod"); err != nil {
			t.Fatal(err)
		}

		loaded, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if loaded.Projects["site"].DefaultEnv != "prod" {
			t.Errorf("persisted DefaultEnv = %q, want %q", loaded.Projects["site"].DefaultEnv, "prod")
		}
	})
}

func TestEnvNames(t *testing.T) {
	t.Run("sorted", func(t *testing.T) {
		reg := &Registry{
			Projects: map[string]*Project{
				"site": {
					Envs: map[string]string{"staging": "/b", "dev": "/a", "production": "/c"},
				},
			},
		}
		names := reg.EnvNames("site")
		want := []string{"dev", "production", "staging"}
		if len(names) != len(want) {
			t.Fatalf("len = %d, want %d", len(names), len(want))
		}
		for i, n := range names {
			if n != want[i] {
				t.Errorf("names[%d] = %q, want %q", i, n, want[i])
			}
		}
	})

	t.Run("missing project returns nil", func(t *testing.T) {
		reg := &Registry{Projects: make(map[string]*Project)}
		names := reg.EnvNames("missing")
		if names != nil {
			t.Errorf("expected nil, got %v", names)
		}
	})

	t.Run("single env", func(t *testing.T) {
		reg := &Registry{
			Projects: map[string]*Project{
				"site": {Envs: map[string]string{"local": "/x"}},
			},
		}
		names := reg.EnvNames("site")
		if len(names) != 1 || names[0] != "local" {
			t.Errorf("expected [local], got %v", names)
		}
	})

	t.Run("empty envs", func(t *testing.T) {
		reg := &Registry{
			Projects: map[string]*Project{
				"site": {Envs: make(map[string]string)},
			},
		}
		names := reg.EnvNames("site")
		if len(names) != 0 {
			t.Errorf("expected empty, got %v", names)
		}
	})
}
