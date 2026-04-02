package main

import (
	"encoding/json"
	"os"

	"github.com/hegner123/modulacms/internal/config"
)

// buildBaseConfig returns a Config with project-level defaults.
// Environment-specific fields (db, credentials, ports) are left at
// zero values since they belong in overlays.
func buildBaseConfig() config.Config {
	c := config.DefaultConfig()

	// Clear environment-specific fields — these go in overlays
	c.Environment = config.Environment("")
	c.Db_Driver = ""
	c.Db_URL = ""
	c.Db_Name = ""
	c.Db_User = ""
	c.Db_Password = ""
	c.Port = ""
	c.SSL_Port = ""
	c.SSH_Port = ""
	c.SSH_Host = ""
	c.Client_Site = ""
	c.Admin_Site = ""
	c.Auth_Salt = ""
	c.Cert_Dir = ""
	c.Log_Path = ""
	c.Bucket_Endpoint = ""
	c.Bucket_Access_Key = ""
	c.Bucket_Secret_Key = ""
	c.Bucket_Public_URL = ""

	return c
}

// buildLocalOverlay returns the local development overlay.
func buildLocalOverlay() map[string]any {
	return map[string]any{
		"environment": "local",
		"db_driver":   "sqlite",
		"db_url":      "./modula.db",
		"db_name":     "modula.db",
		"port":        ":8080",
		"ssl_port":    ":4000",
		"ssh_host":    "localhost",
		"ssh_port":    "2233",
		"client_site": "localhost",
		"admin_site":  "localhost",
		"cert_dir":    "./certs",
		"log_path":    "",
	}
}

// buildDevOverlay returns the dev environment overlay.
func buildDevOverlay() map[string]any {
	return map[string]any{
		"environment": "development",
		"db_driver":   "sqlite",
		"db_url":      "./modula-dev.db",
		"db_name":     "modula-dev.db",
		"port":        ":8080",
		"ssl_port":    ":4000",
		"ssh_host":    "localhost",
		"ssh_port":    "2233",
		"client_site": "localhost",
		"admin_site":  "localhost",
		"cert_dir":    "./certs",
		"log_path":    "",
	}
}

// buildProdOverlay returns the production overlay with env var placeholders.
func buildProdOverlay() map[string]any {
	return map[string]any{
		"environment":    "production",
		"db_driver":      "postgres",
		"db_url":         "${DB_HOST}:5432",
		"db_name":        "${DB_NAME}",
		"db_username":    "${DB_USER}",
		"db_password":    "${DB_PASSWORD}",
		"port":           ":8080",
		"ssl_port":       ":443",
		"ssh_host":       "0.0.0.0",
		"ssh_port":       "2233",
		"client_site":    "${CLIENT_SITE}",
		"admin_site":     "${ADMIN_SITE}",
		"cert_dir":       "${CERT_DIR}",
		"auth_salt":      "${AUTH_SALT}",
		"cookie_secure":  true,
		"bucket_endpoint":   "${BUCKET_ENDPOINT}",
		"bucket_access_key": "${BUCKET_ACCESS_KEY}",
		"bucket_secret_key": "${BUCKET_SECRET_KEY}",
		"bucket_public_url": "${BUCKET_PUBLIC_URL}",
		"log_path":       "/var/log/modulacms",
	}
}

// writeConfigFile marshals a Config struct to indented JSON and writes it.
func writeConfigFile(path string, c config.Config) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0640)
}

// writeOverlayFile marshals a map to indented JSON and writes it.
func writeOverlayFile(path string, data map[string]any) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0640)
}
