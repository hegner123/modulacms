package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// validateScaffoldArgs
// ---------------------------------------------------------------------------

func TestValidateScaffoldArgs(t *testing.T) {
	valid := scaffoldArgs{DBDriver: "sqlite", Storage: "minio", Proxy: "caddy"}
	if err := validateScaffoldArgs(valid); err != nil {
		t.Fatalf("expected valid args, got error: %v", err)
	}

	t.Run("all valid combinations accepted", func(t *testing.T) {
		for _, db := range []string{"sqlite", "mysql", "postgres"} {
			for _, st := range []string{"minio", "external"} {
				for _, px := range []string{"caddy", "nginx", "raw"} {
					a := scaffoldArgs{DBDriver: db, Storage: st, Proxy: px}
					if err := validateScaffoldArgs(a); err != nil {
						t.Errorf("expected valid for db=%s storage=%s proxy=%s, got: %v", db, st, px, err)
					}
				}
			}
		}
	})

	t.Run("invalid db", func(t *testing.T) {
		a := scaffoldArgs{DBDriver: "oracle", Storage: "minio", Proxy: "caddy"}
		err := validateScaffoldArgs(a)
		if err == nil {
			t.Fatal("expected error for invalid db")
		}
		if !strings.Contains(err.Error(), "oracle") {
			t.Errorf("error should mention invalid value, got: %v", err)
		}
	})

	t.Run("invalid storage", func(t *testing.T) {
		a := scaffoldArgs{DBDriver: "sqlite", Storage: "gcs", Proxy: "caddy"}
		err := validateScaffoldArgs(a)
		if err == nil {
			t.Fatal("expected error for invalid storage")
		}
		if !strings.Contains(err.Error(), "gcs") {
			t.Errorf("error should mention invalid value, got: %v", err)
		}
	})

	t.Run("invalid proxy", func(t *testing.T) {
		a := scaffoldArgs{DBDriver: "sqlite", Storage: "minio", Proxy: "traefik"}
		err := validateScaffoldArgs(a)
		if err == nil {
			t.Fatal("expected error for invalid proxy")
		}
		if !strings.Contains(err.Error(), "traefik") {
			t.Errorf("error should mention invalid value, got: %v", err)
		}
	})

	t.Run("empty values rejected", func(t *testing.T) {
		if err := validateScaffoldArgs(scaffoldArgs{DBDriver: "", Storage: "minio", Proxy: "caddy"}); err == nil {
			t.Error("expected error for empty db")
		}
		if err := validateScaffoldArgs(scaffoldArgs{DBDriver: "sqlite", Storage: "", Proxy: "caddy"}); err == nil {
			t.Error("expected error for empty storage")
		}
		if err := validateScaffoldArgs(scaffoldArgs{DBDriver: "sqlite", Storage: "minio", Proxy: ""}); err == nil {
			t.Error("expected error for empty proxy")
		}
	})
}

// ---------------------------------------------------------------------------
// newScaffoldData
// ---------------------------------------------------------------------------

func TestNewScaffoldData(t *testing.T) {
	t.Run("sqlite minio caddy", func(t *testing.T) {
		d := newScaffoldData(scaffoldArgs{DBDriver: "sqlite", Storage: "minio", Proxy: "caddy"})
		if !d.IsSQLite || d.IsMySQL || d.IsPostgres {
			t.Error("db flags wrong")
		}
		if !d.IsMinio || d.IsExternal {
			t.Error("storage flags wrong")
		}
		if !d.IsCaddy || d.IsNginx || d.IsRaw {
			t.Error("proxy flags wrong")
		}
	})

	t.Run("postgres external nginx", func(t *testing.T) {
		d := newScaffoldData(scaffoldArgs{DBDriver: "postgres", Storage: "external", Proxy: "nginx"})
		if d.IsSQLite || d.IsMySQL || !d.IsPostgres {
			t.Error("db flags wrong")
		}
		if d.IsMinio || !d.IsExternal {
			t.Error("storage flags wrong")
		}
		if d.IsCaddy || !d.IsNginx || d.IsRaw {
			t.Error("proxy flags wrong")
		}
	})

	t.Run("mysql minio raw", func(t *testing.T) {
		d := newScaffoldData(scaffoldArgs{DBDriver: "mysql", Storage: "minio", Proxy: "raw"})
		if d.IsSQLite || !d.IsMySQL || d.IsPostgres {
			t.Error("db flags wrong")
		}
		if !d.IsMinio || d.IsExternal {
			t.Error("storage flags wrong")
		}
		if d.IsCaddy || d.IsNginx || !d.IsRaw {
			t.Error("proxy flags wrong")
		}
	})
}

// ---------------------------------------------------------------------------
// Dockerfile
// ---------------------------------------------------------------------------

func TestRenderDockerfile(t *testing.T) {
	out := renderDockerfile(scaffoldData{})

	required := []string{
		"FROM golang:1.25-bookworm AS builder",
		"FROM debian:bookworm-slim",
		"EXPOSE 8080 4000 2233",
		"VOLUME",
		"HEALTHCHECK",
		"ENTRYPOINT",
		"CGO_ENABLED=1",
		"go build -mod vendor",
		"libwebp-dev",
		"USER modula",
	}

	for _, s := range required {
		if !strings.Contains(out, s) {
			t.Errorf("Dockerfile missing %q", s)
		}
	}
}

// ---------------------------------------------------------------------------
// docker-compose.yml — all 18 combinations
// ---------------------------------------------------------------------------

func TestRenderCompose(t *testing.T) {
	// Every combo must have the modula service and core volumes.
	for _, db := range []string{"sqlite", "mysql", "postgres"} {
		for _, st := range []string{"minio", "external"} {
			for _, px := range []string{"caddy", "nginx", "raw"} {
				name := db + "_" + st + "_" + px
				t.Run(name, func(t *testing.T) {
					d := newScaffoldData(scaffoldArgs{DBDriver: db, Storage: st, Proxy: px})
					out := renderCompose(d)

					mustContain(t, out, "modula_cms", "missing CMS container")
					mustContain(t, out, "cms_data:", "missing cms_data volume")
					mustContain(t, out, "cms_ssh:", "missing cms_ssh volume")
					mustContain(t, out, "2233:2233", "missing SSH port")

					// DB service present/absent
					switch db {
					case "postgres":
						mustContain(t, out, "postgres:17", "missing postgres image")
						mustContain(t, out, "pg_data:", "missing pg_data volume")
						mustContain(t, out, "pg_isready", "missing postgres healthcheck")
						mustNotContain(t, out, "mysql:8.0", "postgres should not have mysql")
					case "mysql":
						mustContain(t, out, "mysql:8.0", "missing mysql image")
						mustContain(t, out, "mysql_data:", "missing mysql_data volume")
						mustContain(t, out, "mysqladmin", "missing mysql healthcheck")
						mustNotContain(t, out, "postgres:17", "mysql should not have postgres")
					case "sqlite":
						mustNotContain(t, out, "postgres:17", "sqlite should not have postgres")
						mustNotContain(t, out, "mysql:8.0", "sqlite should not have mysql")
					}

					// Storage service present/absent
					if st == "minio" {
						mustContain(t, out, "minio/minio", "missing minio image")
						mustContain(t, out, "minio_data:", "missing minio_data volume")
						mustContain(t, out, "9001:9001", "missing minio console port")
					} else {
						mustNotContain(t, out, "minio/minio", "external should not have minio service")
						mustNotContain(t, out, "minio_data:", "external should not have minio volume")
					}

					// Proxy present/absent
					switch px {
					case "caddy":
						mustContain(t, out, "caddy:2", "missing caddy image")
						mustContain(t, out, "caddy_data:", "missing caddy_data volume")
						mustContain(t, out, "80:80", "missing http port")
						mustContain(t, out, "443:443", "missing https port")
						mustNotContain(t, out, "nginx:alpine", "caddy should not have nginx")
						mustNotContain(t, out, "8080:8080", "caddy should not expose CMS http directly")
					case "nginx":
						mustContain(t, out, "nginx:alpine", "missing nginx image")
						mustContain(t, out, "80:80", "missing http port")
						mustNotContain(t, out, "caddy:2", "nginx should not have caddy")
						mustNotContain(t, out, "8080:8080", "nginx should not expose CMS http directly")
					case "raw":
						mustContain(t, out, "8080:8080", "raw should expose CMS http port")
						mustContain(t, out, "4000:4000", "raw should expose CMS https port")
						mustNotContain(t, out, "caddy:2", "raw should not have caddy")
						mustNotContain(t, out, "nginx:alpine", "raw should not have nginx")
					}

					// depends_on: DB and minio referenced when present
					if db != "sqlite" || st == "minio" {
						mustContain(t, out, "depends_on:", "missing depends_on")
						mustContain(t, out, "service_healthy", "missing service_healthy condition")
					}
				})
			}
		}
	}
}

// ---------------------------------------------------------------------------
// modula.config.json
// ---------------------------------------------------------------------------

func TestRenderConfig(t *testing.T) {
	t.Run("sqlite minio", func(t *testing.T) {
		d := newScaffoldData(scaffoldArgs{DBDriver: "sqlite", Storage: "minio", Proxy: "raw"})
		out := renderConfig(d)

		mustContain(t, out, `"db_driver": "sqlite"`, "wrong db_driver")
		mustContain(t, out, `"db_url": "/app/data/modula.db"`, "wrong db_url for sqlite")
		mustContain(t, out, `"bucket_endpoint": "minio:9000"`, "wrong bucket endpoint for minio")
		mustContain(t, out, `"bucket_force_path_style": true`, "minio needs path style")

		assertValidJSON(t, out)
	})

	t.Run("postgres external", func(t *testing.T) {
		d := newScaffoldData(scaffoldArgs{DBDriver: "postgres", Storage: "external", Proxy: "caddy"})
		out := renderConfig(d)

		mustContain(t, out, `"db_driver": "postgres"`, "wrong db_driver")
		mustContain(t, out, `"db_url": "postgres"`, "wrong db_url for postgres")
		mustContain(t, out, `"db_name": "modula_db"`, "missing db_name")
		mustContain(t, out, `"bucket_endpoint": "CHANGE_ME_s3_endpoint"`, "wrong bucket endpoint for external")
		mustContain(t, out, `"bucket_force_path_style": false`, "external should not force path style")

		assertValidJSON(t, out)
	})

	t.Run("mysql minio", func(t *testing.T) {
		d := newScaffoldData(scaffoldArgs{DBDriver: "mysql", Storage: "minio", Proxy: "nginx"})
		out := renderConfig(d)

		mustContain(t, out, `"db_driver": "mysql"`, "wrong db_driver")
		mustContain(t, out, `"db_url": "mysql:3306"`, "wrong db_url for mysql")

		assertValidJSON(t, out)
	})

	t.Run("all configs have required fields", func(t *testing.T) {
		for _, db := range []string{"sqlite", "mysql", "postgres"} {
			for _, st := range []string{"minio", "external"} {
				d := newScaffoldData(scaffoldArgs{DBDriver: db, Storage: st, Proxy: "raw"})
				out := renderConfig(d)
				assertValidJSON(t, out)
				mustContain(t, out, `"ssh_port": "2233"`, "missing ssh_port")
				mustContain(t, out, `"plugin_enabled": true`, "missing plugin_enabled")
				mustContain(t, out, `"cookie_secure": true`, "missing cookie_secure")
			}
		}
	})
}

// ---------------------------------------------------------------------------
// .env
// ---------------------------------------------------------------------------

func TestRenderEnv(t *testing.T) {
	t.Run("always has base vars", func(t *testing.T) {
		d := newScaffoldData(scaffoldArgs{DBDriver: "sqlite", Storage: "external", Proxy: "raw"})
		out := renderEnv(d)
		mustContain(t, out, "VERSION=dev", "missing VERSION")
	})

	t.Run("postgres vars", func(t *testing.T) {
		d := newScaffoldData(scaffoldArgs{DBDriver: "postgres", Storage: "external", Proxy: "raw"})
		out := renderEnv(d)
		mustContain(t, out, "POSTGRES_USER=", "missing POSTGRES_USER")
		mustContain(t, out, "POSTGRES_PASSWORD=", "missing POSTGRES_PASSWORD")
		mustContain(t, out, "POSTGRES_DB=", "missing POSTGRES_DB")
		mustNotContain(t, out, "MYSQL_", "postgres env should not have mysql vars")
	})

	t.Run("mysql vars", func(t *testing.T) {
		d := newScaffoldData(scaffoldArgs{DBDriver: "mysql", Storage: "external", Proxy: "raw"})
		out := renderEnv(d)
		mustContain(t, out, "MYSQL_ROOT_PASSWORD=", "missing MYSQL_ROOT_PASSWORD")
		mustContain(t, out, "MYSQL_DATABASE=", "missing MYSQL_DATABASE")
		mustNotContain(t, out, "POSTGRES_", "mysql env should not have postgres vars")
	})

	t.Run("minio vars", func(t *testing.T) {
		d := newScaffoldData(scaffoldArgs{DBDriver: "sqlite", Storage: "minio", Proxy: "raw"})
		out := renderEnv(d)
		mustContain(t, out, "MINIO_ROOT_USER=", "missing MINIO_ROOT_USER")
		mustContain(t, out, "MINIO_ROOT_PASSWORD=", "missing MINIO_ROOT_PASSWORD")
		mustNotContain(t, out, "BUCKET_ENDPOINT=", "minio should not have external S3 vars")
	})

	t.Run("external s3 vars", func(t *testing.T) {
		d := newScaffoldData(scaffoldArgs{DBDriver: "sqlite", Storage: "external", Proxy: "raw"})
		out := renderEnv(d)
		mustContain(t, out, "BUCKET_ENDPOINT=", "missing BUCKET_ENDPOINT")
		mustContain(t, out, "BUCKET_ACCESS_KEY=", "missing BUCKET_ACCESS_KEY")
		mustNotContain(t, out, "MINIO_ROOT_USER=", "external should not have minio vars")
	})

	t.Run("sqlite has no db vars", func(t *testing.T) {
		d := newScaffoldData(scaffoldArgs{DBDriver: "sqlite", Storage: "minio", Proxy: "raw"})
		out := renderEnv(d)
		mustNotContain(t, out, "POSTGRES_", "sqlite should not have postgres vars")
		mustNotContain(t, out, "MYSQL_", "sqlite should not have mysql vars")
	})
}

// ---------------------------------------------------------------------------
// Proxy config files
// ---------------------------------------------------------------------------

func TestRenderCaddyfile(t *testing.T) {
	out := renderCaddyfile(scaffoldData{})
	mustContain(t, out, "reverse_proxy modula:8080", "missing reverse_proxy directive")
	mustContain(t, out, ":80", "missing port 80")
}

func TestRenderNginxConf(t *testing.T) {
	out := renderNginxConf(scaffoldData{})
	mustContain(t, out, "proxy_pass http://modula_backend", "missing proxy_pass")
	mustContain(t, out, "server modula:8080", "missing upstream server")
	mustContain(t, out, "client_max_body_size 100M", "missing client_max_body_size")
	mustContain(t, out, "X-Forwarded-For", "missing forwarded headers")
}

// ---------------------------------------------------------------------------
// File writing
// ---------------------------------------------------------------------------

func TestWriteScaffoldFiles(t *testing.T) {
	dir := t.TempDir()

	a := scaffoldArgs{
		DBDriver: "postgres",
		Storage:  "minio",
		Proxy:    "caddy",
		OutDir:   dir,
	}

	cmd := scaffoldDockerCmd
	cmd.SetOut(&strings.Builder{})

	if err := writeScaffoldFiles(cmd, a); err != nil {
		t.Fatalf("writeScaffoldFiles failed: %v", err)
	}

	expected := []string{"Dockerfile", "docker-compose.yml", "modula.config.json", ".env", "Caddyfile"}
	for _, name := range expected {
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected file %s to exist: %v", name, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("expected file %s to be non-empty", name)
		}
	}

	// nginx.conf should NOT exist for caddy
	if _, err := os.Stat(filepath.Join(dir, "nginx.conf")); err == nil {
		t.Error("nginx.conf should not exist when proxy=caddy")
	}
}

func TestWriteScaffoldFiles_Nginx(t *testing.T) {
	dir := t.TempDir()

	a := scaffoldArgs{
		DBDriver: "sqlite",
		Storage:  "external",
		Proxy:    "nginx",
		OutDir:   dir,
	}

	cmd := scaffoldDockerCmd
	cmd.SetOut(&strings.Builder{})

	if err := writeScaffoldFiles(cmd, a); err != nil {
		t.Fatalf("writeScaffoldFiles failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "nginx.conf")); err != nil {
		t.Error("nginx.conf should exist when proxy=nginx")
	}
	if _, err := os.Stat(filepath.Join(dir, "Caddyfile")); err == nil {
		t.Error("Caddyfile should not exist when proxy=nginx")
	}
}

func TestWriteScaffoldFiles_Raw(t *testing.T) {
	dir := t.TempDir()

	a := scaffoldArgs{
		DBDriver: "sqlite",
		Storage:  "minio",
		Proxy:    "raw",
		OutDir:   dir,
	}

	cmd := scaffoldDockerCmd
	cmd.SetOut(&strings.Builder{})

	if err := writeScaffoldFiles(cmd, a); err != nil {
		t.Fatalf("writeScaffoldFiles failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "nginx.conf")); err == nil {
		t.Error("nginx.conf should not exist when proxy=raw")
	}
	if _, err := os.Stat(filepath.Join(dir, "Caddyfile")); err == nil {
		t.Error("Caddyfile should not exist when proxy=raw")
	}
}

func TestWriteScaffoldFiles_Stdout(t *testing.T) {
	a := scaffoldArgs{
		DBDriver: "postgres",
		Storage:  "minio",
		Proxy:    "caddy",
		Stdout:   true,
	}

	var buf strings.Builder
	cmd := scaffoldDockerCmd
	cmd.SetOut(&buf)

	if err := writeScaffoldFiles(cmd, a); err != nil {
		t.Fatalf("writeScaffoldFiles stdout failed: %v", err)
	}

	out := buf.String()
	mustContain(t, out, "--- Dockerfile ---", "missing Dockerfile header")
	mustContain(t, out, "--- docker-compose.yml ---", "missing compose header")
	mustContain(t, out, "--- modula.config.json ---", "missing config header")
	mustContain(t, out, "--- .env ---", "missing env header")
	mustContain(t, out, "--- Caddyfile ---", "missing Caddyfile header")
	mustNotContain(t, out, "--- nginx.conf ---", "should not have nginx header for caddy")
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func mustContain(t *testing.T, haystack, needle, msg string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("%s: output does not contain %q", msg, needle)
	}
}

func mustNotContain(t *testing.T, haystack, needle, msg string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Errorf("%s: output should not contain %q", msg, needle)
	}
}

func assertValidJSON(t *testing.T, s string) {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Errorf("invalid JSON: %v\n%s", err, s)
	}
}
