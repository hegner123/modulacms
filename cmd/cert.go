package main

import (
	"path/filepath"

	"github.com/hegner123/modulacms/internal/utility"
	"github.com/spf13/cobra"
)

// certCmd is the root command for certificate management operations.
var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "certificate management commands",
	Long: `Manage SSL/TLS certificates for local HTTPS development.

Subcommands:
  generate   Create self-signed certificates for localhost

Examples:
  modula cert generate`,
}

// certGenerateCmd generates self-signed SSL certificates for local development use.
var certGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate self-signed SSL certificates for local development",
	Long: `Generate a self-signed SSL certificate and private key for local HTTPS.

Creates localhost.crt and localhost.key in the cert_dir configured in modula.config.json
(default: ./certs). If a client_site domain is configured, the certificate is
issued for that domain instead of localhost.

After generation, attempts to add the certificate to the system trust store
(macOS Keychain). On failure, prints instructions for manual trust.

To use HTTPS locally, set "environment": "local" in modula.config.json.

Examples:
  modula cert generate
  modula cert generate --config /path/to/modula.config.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		// Root the process in the config directory so relative paths resolve correctly.
		resolveConfigDir()

		utility.DefaultLogger.Info("generating self-signed SSL certificates...")

		// Defaults
		certDir := "./certs"
		domain := "localhost"

		// Try to read config for cert_dir and domain settings (graceful fallback)
		mgr, loadErr := loadConfig()
		if loadErr == nil {
			if cfg, cfgErr := mgr.Config(); cfgErr == nil {
				if cfg.Cert_Dir != "" {
					certDir = cfg.Cert_Dir
				}
				if cfg.Client_Site != "" && cfg.Client_Site != "localhost" && cfg.Client_Site != "0.0.0.0" && cfg.Client_Site != "127.0.0.1" {
					domain = cfg.Client_Site
				}
			}
		}

		utility.DefaultLogger.Info("certificate directory", certDir)
		utility.DefaultLogger.Info("domain", domain)

		if err := utility.GenerateSelfSignedCert(certDir, domain); err != nil {
			return err
		}

		utility.DefaultLogger.Info("successfully generated SSL certificates")
		utility.DefaultLogger.Info("  - Certificate: " + certDir + "/localhost.crt")
		utility.DefaultLogger.Info("  - Private Key: " + certDir + "/localhost.key")

		// Offer to trust the certificate
		certPath := filepath.Join(certDir, "localhost.crt")
		if err := utility.TrustCertificate(certPath); err != nil {
			utility.DefaultLogger.Warn("failed to trust certificate:", err)
			utility.DefaultLogger.Info("you can manually trust it later - see LOCAL_HTTPS_SETUP.md")
		}

		utility.DefaultLogger.Info("to use HTTPS locally, set environment to 'local' in modula.config.json")
		return nil
	},
}

func init() {
	certCmd.AddCommand(certGenerateCmd)
}
