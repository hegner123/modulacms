package install_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/install"
)

func TestCheckConfigExists(t *testing.T) {
	t.Parallel()

	t.Run("existing file returns nil", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "modula.config.json")
		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
		err := install.CheckConfigExists(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("nonexistent file returns error", func(t *testing.T) {
		t.Parallel()
		err := install.CheckConfigExists("/nonexistent/modula.config.json")
		if err == nil {
			t.Fatal("expected error for nonexistent config, got nil")
		}
	})
}

func TestCheckCerts(t *testing.T) {
	t.Parallel()

	t.Run("both cert and key exist returns true", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "localhost.crt"), []byte("cert"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "localhost.key"), []byte("key"), 0644); err != nil {
			t.Fatal(err)
		}
		if !install.CheckCerts(dir) {
			t.Error("expected true when both cert files exist")
		}
	})

	t.Run("missing cert returns false", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "localhost.key"), []byte("key"), 0644); err != nil {
			t.Fatal(err)
		}
		if install.CheckCerts(dir) {
			t.Error("expected false when cert file is missing")
		}
	})

	t.Run("missing key returns false", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "localhost.crt"), []byte("cert"), 0644); err != nil {
			t.Fatal(err)
		}
		if install.CheckCerts(dir) {
			t.Error("expected false when key file is missing")
		}
	})

	t.Run("empty directory returns false", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		if install.CheckCerts(dir) {
			t.Error("expected false for empty directory")
		}
	})

	t.Run("nonexistent directory returns false", func(t *testing.T) {
		t.Parallel()
		if install.CheckCerts("/nonexistent/certs") {
			t.Error("expected false for nonexistent directory")
		}
	})
}

func TestCheckOauth(t *testing.T) {
	t.Parallel()

	t.Run("empty credentials returns not configured", func(t *testing.T) {
		t.Parallel()
		c := &config.Config{}
		status, err := install.CheckOauth(nil, c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != "Not configured" {
			t.Errorf("status = %q, want %q", status, "Not configured")
		}
	})

	t.Run("partial credentials returns not configured", func(t *testing.T) {
		t.Parallel()
		c := &config.Config{
			Oauth_Client_Id: "some-id",
			// missing secret and endpoints
		}
		status, err := install.CheckOauth(nil, c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != "Not configured" {
			t.Errorf("status = %q, want %q", status, "Not configured")
		}
	})

	t.Run("full credentials returns connected", func(t *testing.T) {
		t.Parallel()
		c := &config.Config{
			Oauth_Client_Id:     "client-id",
			Oauth_Client_Secret: "client-secret",
			Oauth_Endpoint: map[config.Endpoint]string{
				config.OauthAuthURL:  "https://auth.example.com/authorize",
				config.OauthTokenURL: "https://auth.example.com/token",
			},
		}
		status, err := install.CheckOauth(nil, c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != "Connected" {
			t.Errorf("status = %q, want %q", status, "Connected")
		}
	})

	t.Run("verbose flag does not change result", func(t *testing.T) {
		t.Parallel()
		v := true
		c := &config.Config{}
		status, err := install.CheckOauth(&v, c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != "Not configured" {
			t.Errorf("status = %q, want %q", status, "Not configured")
		}
	})

	t.Run("missing auth url is not configured", func(t *testing.T) {
		t.Parallel()
		c := &config.Config{
			Oauth_Client_Id:     "id",
			Oauth_Client_Secret: "secret",
			Oauth_Endpoint: map[config.Endpoint]string{
				config.OauthTokenURL: "https://auth.example.com/token",
				// missing oauth_auth_url
			},
		}
		status, err := install.CheckOauth(nil, c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != "Not configured" {
			t.Errorf("status = %q, want %q", status, "Not configured")
		}
	})
}

func TestCheckBucket_EmptyCredentials(t *testing.T) {
	t.Parallel()

	c := &config.Config{}
	status, err := install.CheckBucket(nil, c)
	if err != nil {
		t.Fatalf("unexpected error for empty credentials: %v", err)
	}
	if status != "Not configured" {
		t.Errorf("status = %q, want %q", status, "Not configured")
	}
}

func TestCheckBucket_PartialCredentials(t *testing.T) {
	t.Parallel()

	c := &config.Config{
		Bucket_Access_Key: "key",
		// missing secret and endpoint
	}
	status, err := install.CheckBucket(nil, c)
	if err != nil {
		t.Fatalf("unexpected error for partial credentials: %v", err)
	}
	if status != "Not configured" {
		t.Errorf("status = %q, want %q", status, "Not configured")
	}
}

func TestDBStatus_Fields(t *testing.T) {
	t.Parallel()

	s := install.DBStatus{
		Driver: "sqlite",
		URL:    "/data/test.db",
		Err:    nil,
	}
	if s.Driver != "sqlite" {
		t.Errorf("Driver = %q, want %q", s.Driver, "sqlite")
	}
	if s.URL != "/data/test.db" {
		t.Errorf("URL = %q, want %q", s.URL, "/data/test.db")
	}
	if s.Err != nil {
		t.Errorf("Err = %v, want nil", s.Err)
	}
}
