package main

import (
	"path/filepath"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/spf13/cobra"
)

// certCmd is the root command for certificate management operations.
var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "Certificate management commands",
}

// certGenerateCmd generates self-signed SSL certificates for local development use.
var certGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate self-signed SSL certificates for local development",
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		utility.DefaultLogger.Info("Generating self-signed SSL certificates...")

		// Defaults
		certDir := "./certs"
		domain := "localhost"

		// Try to read config for cert_dir and domain settings (graceful fallback)
		configProvider := config.NewFileProvider(cfgPath)
		configManager := config.NewManager(configProvider)
		if err := configManager.Load(); err == nil {
			if cfg, err := configManager.Config(); err == nil {
				if cfg.Cert_Dir != "" {
					certDir = cfg.Cert_Dir
				}
				if cfg.Client_Site != "" && cfg.Client_Site != "localhost" && cfg.Client_Site != "0.0.0.0" && cfg.Client_Site != "127.0.0.1" {
					domain = cfg.Client_Site
				}
			}
		}

		utility.DefaultLogger.Info("Certificate directory", certDir)
		utility.DefaultLogger.Info("Domain", domain)

		if err := utility.GenerateSelfSignedCert(certDir, domain); err != nil {
			return err
		}

		utility.DefaultLogger.Info("Successfully generated SSL certificates")
		utility.DefaultLogger.Info("  - Certificate: " + certDir + "/localhost.crt")
		utility.DefaultLogger.Info("  - Private Key: " + certDir + "/localhost.key")

		// Offer to trust the certificate
		certPath := filepath.Join(certDir, "localhost.crt")
		if err := utility.TrustCertificate(certPath); err != nil {
			utility.DefaultLogger.Warn("Failed to trust certificate:", err)
			utility.DefaultLogger.Info("You can manually trust it later - see LOCAL_HTTPS_SETUP.md")
		}

		utility.DefaultLogger.Info("To use HTTPS locally, set environment to 'local' in config.json")
		return nil
	},
}

func init() {
	certCmd.AddCommand(certGenerateCmd)
}
