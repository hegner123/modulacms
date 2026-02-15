package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/plugin"
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

// initPluginPool creates an independent database connection pool for plugin use.
// Returns the pool and a cleanup function that closes it.
func initPluginPool(cfg *config.Config) (*sql.DB, func(), error) {
	pc := db.DefaultPluginPoolConfig()

	if cfg.Plugin_DB_MaxOpenConns > 0 {
		pc.MaxOpenConns = cfg.Plugin_DB_MaxOpenConns
	}
	if cfg.Plugin_DB_MaxIdleConns > 0 {
		pc.MaxIdleConns = cfg.Plugin_DB_MaxIdleConns
	}
	if cfg.Plugin_DB_ConnMaxLifetime != "" {
		d, err := time.ParseDuration(cfg.Plugin_DB_ConnMaxLifetime)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid plugin_db_conn_max_lifetime %q: %w", cfg.Plugin_DB_ConnMaxLifetime, err)
		}
		pc.ConnMaxLifetime = d
	}

	pool, err := db.OpenPool(*cfg, pc)
	if err != nil {
		return nil, nil, fmt.Errorf("plugin pool: %w", err)
	}

	utility.DefaultLogger.Info("Plugin database pool initialized",
		"max_open", pc.MaxOpenConns,
		"max_idle", pc.MaxIdleConns,
		"max_lifetime", pc.ConnMaxLifetime,
	)

	cleanup := func() {
		utility.DefaultLogger.Info("Closing plugin database pool")
		if cerr := pool.Close(); cerr != nil {
			utility.DefaultLogger.Error("Plugin pool close error", cerr)
		}
	}

	return pool, cleanup, nil
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

// initPluginManager creates and starts the plugin Manager if plugin_enabled is true.
// Returns nil when plugins are disabled. The caller must defer Manager.Shutdown()
// when the returned manager is non-nil.
//
// Note: Manager.Shutdown() also closes the *sql.DB pool it receives. Since
// pluginPoolCleanup (registered earlier in the defer stack) will attempt to close
// the same pool afterward, the second close is a benign no-op that logs an error.
// This is intentional — keeping both defers avoids a nil-pool panic if
// initPluginManager returns nil but the pool was already created.
func initPluginManager(ctx context.Context, cfg *config.Config, pool *sql.DB) *plugin.Manager {
	if !cfg.Plugin_Enabled {
		utility.DefaultLogger.Info("plugin system disabled")
		return nil
	}

	dir := cfg.Plugin_Directory
	if dir == "" {
		dir = "./plugins/"
	}

	// Map config driver name to query builder dialect.
	// config.DbDriver values ("sqlite", "mysql", "postgres") are accepted by
	// db.DialectFromString; unrecognized values default to DialectSQLite.
	dialect := db.DialectFromString(string(cfg.Db_Driver))

	// Phase 4: Parse circuit breaker reset interval.
	maxFailures := cfg.Plugin_Max_Failures
	if maxFailures <= 0 {
		maxFailures = 5
	}
	resetInterval := 60 * time.Second
	if cfg.Plugin_Reset_Interval != "" {
		parsed, parseErr := time.ParseDuration(cfg.Plugin_Reset_Interval)
		if parseErr != nil {
			utility.DefaultLogger.Warn(
				fmt.Sprintf("invalid plugin_reset_interval %q, using default 60s: %s",
					cfg.Plugin_Reset_Interval, parseErr.Error()),
				nil,
			)
		} else {
			resetInterval = parsed
		}
	}

	mgr := plugin.NewManager(plugin.ManagerConfig{
		Enabled:         cfg.Plugin_Enabled,
		Directory:       dir,
		MaxVMsPerPlugin: cfg.Plugin_Max_VMs,
		ExecTimeoutSec:  cfg.Plugin_Timeout,
		MaxOpsPerExec:   cfg.Plugin_Max_Ops,

		// Phase 3: Hook engine configuration.
		HookReserveVMs:          cfg.Plugin_Hook_Reserve_VMs,
		HookMaxConsecutiveAborts: cfg.Plugin_Hook_Max_Consecutive_Aborts,
		HookMaxOps:              cfg.Plugin_Hook_Max_Ops,
		HookMaxConcurrentAfter:  cfg.Plugin_Hook_Max_Concurrent_After,
		HookTimeoutMs:           cfg.Plugin_Hook_Timeout_Ms,
		HookEventTimeoutMs:      cfg.Plugin_Hook_Event_Timeout_Ms,

		// Phase 4: Production hardening.
		HotReload:     cfg.Plugin_Hot_Reload,
		MaxFailures:   maxFailures,
		ResetInterval: resetInterval,
	}, pool, dialect)

	// Create HTTP bridge and wire it before LoadAll so that LoadAll can
	// register plugin routes and manage the plugin_routes table.
	bridge := plugin.NewHTTPBridge(mgr, pool, dialect)
	mgr.SetBridge(bridge)

	if err := mgr.LoadAll(ctx); err != nil {
		utility.DefaultLogger.Error("plugin system load error", err)
	}

	// Phase 4: Start hot reload watcher if enabled.
	if cfg.Plugin_Hot_Reload {
		mgr.StartWatcher(ctx)
	}

	return mgr
}
