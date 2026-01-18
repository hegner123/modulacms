package flags

import "flag"

type AppFlags struct {
	AuthFlag     *bool
	UpdateFlag   *bool
	CliFlag      *bool
	VersionFlag  *bool
	AlphaFlag    *bool
	VerboseFlag  *bool
	ResetFlag    *bool
	InstallFlag  *bool
	GenCertsFlag *bool
	ConfigPath   *string
}

func ParseFlags() AppFlags {
	app := &AppFlags{
		AuthFlag:     new(bool),
		UpdateFlag:   new(bool),
		CliFlag:      new(bool),
		VersionFlag:  new(bool),
		AlphaFlag:    new(bool),
		VerboseFlag:  new(bool),
		ResetFlag:    new(bool),
		InstallFlag:  new(bool),
		GenCertsFlag: new(bool),
		ConfigPath:   new(string),
	}
	flag.BoolVar(app.AuthFlag, "auth", false, "Run oauth tests")
	flag.BoolVar(app.UpdateFlag, "update", false, "Update binaries and plugins.")
	flag.BoolVar(app.CliFlag, "cli", false, "Launch the Cli without the server.")
	flag.BoolVar(app.VersionFlag, "version", false, "Print version and exit")
	flag.BoolVar(app.AlphaFlag, "a", false, "including code for build purposes")
	flag.BoolVar(app.VerboseFlag, "v", false, "Enable verbose mode")
	flag.BoolVar(app.ResetFlag, "reset", false, "Delete Database and reinitialize")
	flag.BoolVar(app.InstallFlag, "i", false, "Run Installation UI")
	flag.BoolVar(app.GenCertsFlag, "gen-certs", false, "Generate self-signed SSL certificates for local development")
	flag.StringVar(app.ConfigPath, "config", "config.json", "Path to configuration file")
	flag.Parse()
	return *app
}
