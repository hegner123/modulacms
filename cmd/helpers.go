package main

import (
	"context"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

// configureLogger sets the logger level based on the verbose flag.
func configureLogger() {
	if verbose {
		utility.DefaultLogger.SetLevel(utility.DEBUG)
	} else {
		utility.DefaultLogger.SetLevel(utility.INFO)
	}
}

// loadConfig loads the configuration from cfgPath.
func loadConfig() (*config.Config, error) {
	configProvider := config.NewFileProvider(cfgPath)
	configManager := config.NewManager(configProvider)

	if err := configManager.Load(); err != nil {
		return nil, err
	}

	return configManager.Config()
}

// loadConfigAndDB loads the configuration and initializes the singleton database pool.
// The caller is responsible for calling db.CloseDB() when done.
func loadConfigAndDB() (*config.Config, db.DbDriver, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, nil, err
	}

	driver, err := db.InitDB(*cfg)
	if err != nil {
		return nil, nil, err
	}

	return cfg, driver, nil
}

// initObservability starts the observability client if enabled.
// Returns a cleanup function that should be deferred.
func initObservability(ctx context.Context, cfg *config.Config) func() {
	noop := func() {}

	if !cfg.Observability_Enabled {
		return noop
	}

	obsClient, err := utility.NewObservabilityClient(utility.ObservabilityConfig{
		Enabled:       cfg.Observability_Enabled,
		Provider:      cfg.Observability_Provider,
		DSN:           cfg.Observability_DSN,
		Environment:   cfg.Observability_Environment,
		Release:       cfg.Observability_Release,
		SampleRate:    cfg.Observability_Sample_Rate,
		TracesRate:    cfg.Observability_Traces_Rate,
		SendPII:       cfg.Observability_Send_PII,
		Debug:         cfg.Observability_Debug,
		ServerName:    cfg.Observability_Server_Name,
		FlushInterval: cfg.Observability_Flush_Interval,
		Tags:          cfg.Observability_Tags,
	})
	if err != nil {
		utility.DefaultLogger.Error("Failed to initialize observability", err)
		return noop
	}

	utility.GlobalObservability = obsClient
	obsClient.Start(ctx)

	utility.DefaultLogger.Info("Observability started",
		"provider", cfg.Observability_Provider,
		"environment", cfg.Observability_Environment,
		"interval", cfg.Observability_Flush_Interval,
	)

	return func() {
		utility.DefaultLogger.Info("Stopping observability...")
		if err := obsClient.Stop(); err != nil {
			utility.DefaultLogger.Error("Observability shutdown error", err)
		}
	}
}

// logConfigSummary logs the loaded configuration details.
func logConfigSummary(cfg *config.Config) {
	utility.DefaultLogger.Info("Configuration loaded successfully")
	utility.DefaultLogger.Info("Database", "driver", cfg.Db_Driver, "url", cfg.Db_URL)
	utility.DefaultLogger.Info("Sites", "client", cfg.Client_Site, "admin", cfg.Admin_Site)
	utility.DefaultLogger.Info("Ports", "http", cfg.Port, "https", cfg.SSL_Port, "ssh", cfg.SSH_Port)
	utility.DefaultLogger.Info("Environment", "env", cfg.Environment, "host", cfg.Environment_Hosts[cfg.Environment])
	if cfg.Oauth_Provider_Name != "" {
		utility.DefaultLogger.Info("OAuth", "provider", cfg.Oauth_Provider_Name, "redirect", cfg.Oauth_Redirect_URL)
	}
	if cfg.Bucket_Endpoint != "" {
		utility.DefaultLogger.Info("Storage", "endpoint", cfg.Bucket_Endpoint, "media", cfg.Bucket_Media, "backup", cfg.Bucket_Backup)
	}
	utility.DefaultLogger.Info("CORS", "origins", cfg.Cors_Origins, "credentials", cfg.Cors_Credentials)
}
