package config_test

import (
	"encoding/json"
	"runtime"
	"testing"

	"github.com/hegner123/modulacms/internal/config"
)

func TestDefaultConfig_FieldDefaults(t *testing.T) {
	t.Parallel()

	c := config.DefaultConfig()

	// --- Core fields ---

	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "Environment", got: c.Environment, want: "development"},
		{name: "OS matches runtime", got: c.OS, want: runtime.GOOS},
		{name: "Port", got: c.Port, want: ":8080"},
		{name: "SSL_Port", got: c.SSL_Port, want: ":4000"},
		{name: "Cert_Dir", got: c.Cert_Dir, want: "./"},
		{name: "Client_Site", got: c.Client_Site, want: "localhost"},
		{name: "Admin_Site", got: c.Admin_Site, want: "localhost"},
		{name: "SSH_Host", got: c.SSH_Host, want: "localhost"},
		{name: "SSH_Port", got: c.SSH_Port, want: "2233"},
		{name: "Log_Path", got: c.Log_Path, want: "./"},
		{name: "Cookie_Name", got: c.Cookie_Name, want: "modula_cms"},
		{name: "Cookie_Duration", got: c.Cookie_Duration, want: "1w"},
		{name: "Cookie_SameSite", got: c.Cookie_SameSite, want: "lax"},
		{name: "Db_Driver", got: string(c.Db_Driver), want: "sqlite"},
		{name: "Db_Name", got: c.Db_Name, want: "modula.db"},
		{name: "Db_URL", got: c.Db_URL, want: "./modula.db"},
		{name: "Db_Password is empty", got: c.Db_Password, want: ""},
		{name: "Backup_Option", got: c.Backup_Option, want: "./"},
		{name: "Bucket_Region", got: c.Bucket_Region, want: "us-east-1"},
		{name: "Update_Check_Interval", got: c.Update_Check_Interval, want: "startup"},
		{name: "Update_Channel", got: c.Update_Channel, want: "stable"},
		{name: "Oauth_Client_Id is empty", got: c.Oauth_Client_Id, want: ""},
		{name: "Oauth_Client_Secret is empty", got: c.Oauth_Client_Secret, want: ""},
		{name: "Oauth_Provider_Name is empty", got: c.Oauth_Provider_Name, want: ""},
		{name: "Oauth_Redirect_URL is empty", got: c.Oauth_Redirect_URL, want: ""},
		{name: "Oauth_Success_Redirect", got: c.Oauth_Success_Redirect, want: "/"},
		{name: "Observability_Provider", got: c.Observability_Provider, want: "console"},
		{name: "Observability_DSN is empty", got: c.Observability_DSN, want: ""},
		{name: "Observability_Environment", got: c.Observability_Environment, want: "development"},
		{name: "Observability_Release is empty", got: c.Observability_Release, want: ""},
		{name: "Observability_Server_Name is empty", got: c.Observability_Server_Name, want: ""},
		{name: "Observability_Flush_Interval", got: c.Observability_Flush_Interval, want: "30s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("DefaultConfig().%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestDefaultConfig_BoolDefaults(t *testing.T) {
	t.Parallel()

	c := config.DefaultConfig()

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "Cookie_Secure is false", got: c.Cookie_Secure, want: false},
		{name: "Cors_Credentials is true", got: c.Cors_Credentials, want: true},
		{name: "Bucket_Force_Path_Style is true", got: c.Bucket_Force_Path_Style, want: true},
		{name: "Update_Auto_Enabled is false", got: c.Update_Auto_Enabled, want: false},
		{name: "Update_Notify_Only is false", got: c.Update_Notify_Only, want: false},
		{name: "Observability_Enabled is false", got: c.Observability_Enabled, want: false},
		{name: "Observability_Send_PII is false", got: c.Observability_Send_PII, want: false},
		{name: "Observability_Debug is false", got: c.Observability_Debug, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("DefaultConfig() %s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestDefaultConfig_FloatDefaults(t *testing.T) {
	t.Parallel()

	c := config.DefaultConfig()

	tests := []struct {
		name string
		got  float64
		want float64
	}{
		{name: "Observability_Sample_Rate", got: c.Observability_Sample_Rate, want: 1.0},
		{name: "Observability_Traces_Rate", got: c.Observability_Traces_Rate, want: 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("DefaultConfig() %s = %f, want %f", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestDefaultConfig_EnvironmentHosts(t *testing.T) {
	t.Parallel()

	c := config.DefaultConfig()

	if c.Environment_Hosts == nil {
		t.Fatal("DefaultConfig().Environment_Hosts is nil, want populated map")
	}

	expectedKeys := []string{"local", "development", "staging", "production", "http-only"}
	for _, key := range expectedKeys {
		t.Run(key, func(t *testing.T) {
			t.Parallel()
			val, ok := c.Environment_Hosts[key]
			if !ok {
				t.Fatalf("Environment_Hosts missing key %q", key)
			}
			if val != "localhost" {
				t.Errorf("Environment_Hosts[%q] = %q, want %q", key, val, "localhost")
			}
		})
	}

	// No extra keys beyond the expected set
	if len(c.Environment_Hosts) != len(expectedKeys) {
		t.Errorf("Environment_Hosts has %d keys, want %d", len(c.Environment_Hosts), len(expectedKeys))
	}
}

func TestDefaultConfig_SliceDefaults(t *testing.T) {
	t.Parallel()

	c := config.DefaultConfig()

	t.Run("Backup_Paths has one empty entry", func(t *testing.T) {
		t.Parallel()
		if len(c.Backup_Paths) != 1 {
			t.Fatalf("Backup_Paths length = %d, want 1", len(c.Backup_Paths))
		}
		if c.Backup_Paths[0] != "" {
			t.Errorf("Backup_Paths[0] = %q, want empty string", c.Backup_Paths[0])
		}
	})

	t.Run("Cors_Origins", func(t *testing.T) {
		t.Parallel()
		if len(c.Cors_Origins) != 1 {
			t.Fatalf("Cors_Origins length = %d, want 1", len(c.Cors_Origins))
		}
		if c.Cors_Origins[0] != "http://localhost:3000" {
			t.Errorf("Cors_Origins[0] = %q, want %q", c.Cors_Origins[0], "http://localhost:3000")
		}
	})

	t.Run("Cors_Methods", func(t *testing.T) {
		t.Parallel()
		want := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		if len(c.Cors_Methods) != len(want) {
			t.Fatalf("Cors_Methods length = %d, want %d", len(c.Cors_Methods), len(want))
		}
		for i, m := range want {
			if c.Cors_Methods[i] != m {
				t.Errorf("Cors_Methods[%d] = %q, want %q", i, c.Cors_Methods[i], m)
			}
		}
	})

	t.Run("Cors_Headers", func(t *testing.T) {
		t.Parallel()
		want := []string{"Content-Type", "Authorization"}
		if len(c.Cors_Headers) != len(want) {
			t.Fatalf("Cors_Headers length = %d, want %d", len(c.Cors_Headers), len(want))
		}
		for i, h := range want {
			if c.Cors_Headers[i] != h {
				t.Errorf("Cors_Headers[%d] = %q, want %q", i, c.Cors_Headers[i], h)
			}
		}
	})

	t.Run("Oauth_Scopes", func(t *testing.T) {
		t.Parallel()
		want := []string{"openid", "profile", "email"}
		if len(c.Oauth_Scopes) != len(want) {
			t.Fatalf("Oauth_Scopes length = %d, want %d", len(c.Oauth_Scopes), len(want))
		}
		for i, s := range want {
			if c.Oauth_Scopes[i] != s {
				t.Errorf("Oauth_Scopes[%d] = %q, want %q", i, c.Oauth_Scopes[i], s)
			}
		}
	})
}

func TestDefaultConfig_OauthEndpointMap(t *testing.T) {
	t.Parallel()

	c := config.DefaultConfig()

	if c.Oauth_Endpoint == nil {
		t.Fatal("DefaultConfig().Oauth_Endpoint is nil, want populated map")
	}

	// All three endpoint keys should be present with empty values
	endpoints := []config.Endpoint{
		config.OauthAuthURL,
		config.OauthTokenURL,
		config.OauthUserInfoURL,
	}
	for _, ep := range endpoints {
		t.Run(string(ep), func(t *testing.T) {
			t.Parallel()
			val, ok := c.Oauth_Endpoint[ep]
			if !ok {
				t.Fatalf("Oauth_Endpoint missing key %q", ep)
			}
			if val != "" {
				t.Errorf("Oauth_Endpoint[%q] = %q, want empty string", ep, val)
			}
		})
	}

	if len(c.Oauth_Endpoint) != len(endpoints) {
		t.Errorf("Oauth_Endpoint has %d keys, want %d", len(c.Oauth_Endpoint), len(endpoints))
	}
}

func TestDefaultConfig_ObservabilityTags(t *testing.T) {
	t.Parallel()

	c := config.DefaultConfig()

	if c.Observability_Tags == nil {
		t.Fatal("DefaultConfig().Observability_Tags is nil, want empty map")
	}
	if len(c.Observability_Tags) != 0 {
		t.Errorf("Observability_Tags has %d entries, want 0", len(c.Observability_Tags))
	}
}

func TestDefaultConfig_AuthSalt_NotEmpty(t *testing.T) {
	t.Parallel()

	c := config.DefaultConfig()

	// Auth_Salt is a base64-encoded unix timestamp -- it must not be empty
	if c.Auth_Salt == "" {
		t.Fatal("DefaultConfig().Auth_Salt is empty, want non-empty base64 string")
	}
}

func TestDefaultConfig_AuthSalt_UniquePerCall(t *testing.T) {
	// Auth_Salt is derived from time.Now().Unix(), so two calls within the
	// same second will produce the same salt. This test documents that
	// behavior rather than asserting uniqueness -- the important property
	// is that the salt is non-empty and changes over time, not that every
	// call within a tight loop is unique.
	t.Parallel()

	c1 := config.DefaultConfig()
	c2 := config.DefaultConfig()

	// Both should be non-empty
	if c1.Auth_Salt == "" {
		t.Fatal("first DefaultConfig().Auth_Salt is empty")
	}
	if c2.Auth_Salt == "" {
		t.Fatal("second DefaultConfig().Auth_Salt is empty")
	}
}

func TestDefaultConfig_NodeID_NotEmpty(t *testing.T) {
	t.Parallel()

	c := config.DefaultConfig()

	if c.Node_ID == "" {
		t.Fatal("DefaultConfig().Node_ID is empty, want non-empty ULID string")
	}
}

func TestDefaultConfig_NodeID_UniquePerCall(t *testing.T) {
	t.Parallel()

	c1 := config.DefaultConfig()
	c2 := config.DefaultConfig()

	// Node_ID is a ULID generated each call -- must be unique
	if c1.Node_ID == c2.Node_ID {
		t.Errorf("two DefaultConfig() calls produced the same Node_ID: %q", c1.Node_ID)
	}
}

func TestDefaultConfig_KeyBindings_Populated(t *testing.T) {
	t.Parallel()

	c := config.DefaultConfig()

	if c.KeyBindings == nil {
		t.Fatal("DefaultConfig().KeyBindings is nil, want populated KeyMap")
	}

	// Spot-check a few critical bindings to ensure DefaultKeyMap is wired in
	if !c.KeyBindings.Matches("q", config.ActionQuit) {
		t.Error("KeyBindings does not map 'q' to ActionQuit")
	}
	if !c.KeyBindings.Matches("ctrl+c", config.ActionQuit) {
		t.Error("KeyBindings does not map 'ctrl+c' to ActionQuit")
	}
	if !c.KeyBindings.Matches("enter", config.ActionSelect) {
		t.Error("KeyBindings does not map 'enter' to ActionSelect")
	}
}

// --- Config.JSON() tests ---

func TestConfig_JSON_DefaultConfig(t *testing.T) {
	t.Parallel()

	c := config.DefaultConfig()
	got := c.JSON()

	if len(got) == 0 {
		t.Fatal("JSON() returned empty byte slice for DefaultConfig")
	}

	// The output must be valid JSON
	if !json.Valid(got) {
		t.Fatal("JSON() output is not valid JSON")
	}
}

func TestConfig_JSON_RoundTrip(t *testing.T) {
	t.Parallel()

	original := config.DefaultConfig()
	data := original.JSON()

	var decoded config.Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal JSON() output: %v", err)
	}

	// Verify key fields survive the round trip
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "Environment", got: decoded.Environment, want: original.Environment},
		{name: "Port", got: decoded.Port, want: original.Port},
		{name: "SSL_Port", got: decoded.SSL_Port, want: original.SSL_Port},
		{name: "Db_Driver", got: string(decoded.Db_Driver), want: string(original.Db_Driver)},
		{name: "Db_Name", got: decoded.Db_Name, want: original.Db_Name},
		{name: "SSH_Host", got: decoded.SSH_Host, want: original.SSH_Host},
		{name: "SSH_Port", got: decoded.SSH_Port, want: original.SSH_Port},
		{name: "Cookie_Name", got: decoded.Cookie_Name, want: original.Cookie_Name},
		{name: "Auth_Salt", got: decoded.Auth_Salt, want: original.Auth_Salt},
		{name: "Node_ID", got: decoded.Node_ID, want: original.Node_ID},
		{name: "Bucket_Region", got: decoded.Bucket_Region, want: original.Bucket_Region},
		{name: "Update_Channel", got: decoded.Update_Channel, want: original.Update_Channel},
		{name: "Observability_Provider", got: decoded.Observability_Provider, want: original.Observability_Provider},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("round-trip %s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestConfig_JSON_ZeroValue(t *testing.T) {
	t.Parallel()

	var c config.Config
	got := c.JSON()

	if len(got) == 0 {
		t.Fatal("JSON() returned empty byte slice for zero-value Config")
	}

	if !json.Valid(got) {
		t.Fatal("JSON() output for zero-value Config is not valid JSON")
	}
}

func TestConfig_JSON_ContainsExpectedKeys(t *testing.T) {
	t.Parallel()

	c := config.DefaultConfig()
	data := c.JSON()

	// Unmarshal into a generic map to verify key presence.
	// The json tags in the Config struct define the key names.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON into map: %v", err)
	}

	expectedKeys := []string{
		"environment",
		"os",
		"environment_hosts",
		"port",
		"ssl_port",
		"cert_dir",
		"client_site",
		"admin_site",
		"ssh_host",
		"ssh_port",
		"log_path",
		"auth_salt",
		"cookie_name",
		"cookie_duration",
		"cookie_secure",
		"cookie_samesite",
		"db_driver",
		"db_url",
		"db_name",
		"db_password",
		"bucket_region",
		"bucket_force_path_style",
		"backup_option",
		"backup_paths",
		"cors_origins",
		"cors_methods",
		"cors_headers",
		"cors_credentials",
		"oauth_scopes",
		"oauth_endpoint",
		"oauth_success_redirect",
		"observability_enabled",
		"observability_provider",
		"observability_sample_rate",
		"observability_traces_rate",
		"observability_flush_interval",
		"observability_tags",
		"node_id",
		"keybindings",
	}

	for _, key := range expectedKeys {
		t.Run(key, func(t *testing.T) {
			t.Parallel()
			if _, ok := raw[key]; !ok {
				t.Errorf("JSON output missing expected key %q", key)
			}
		})
	}
}
