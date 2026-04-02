# CLI Reference

Complete reference for all `modula` commands.

## Global Flags

These flags apply to every command:

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `./modula.config.json` | Path to config file |
| `--overlay` | | Overlay config file (merged on top of `--config`) |
| `--verbose`, `-v` | `false` | Enable debug-level log output |
| `--yes`, `-y` | `false` | Auto-accept all prompts (maps to `--mode ci` for init) |

## Project Registry

ModulaCMS maintains a project registry at `~/.modula/configs.json` that maps project names to environments and their config file paths. Commands that accept `[project] [environment]` positional args resolve the config path from this registry instead of using `--config`.

```bash
modula serve mysite              # use mysite's default environment
modula serve mysite production   # use mysite's production environment
modula tui mysite staging        # same resolution for tui
```

An explicit `--config` flag always takes priority over positional args. If neither is given, the command looks for `./modula.config.json` in the current directory.

Manage the registry with `modula connect`:

```bash
modula connect set mysite local ./modula.config.json
modula connect set mysite production /etc/modula/prod.config.json
modula connect default mysite
modula connect list
```

---

## Commands

### serve

Start the HTTP, HTTPS, and SSH servers.

```bash
modula serve [project] [environment]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--wizard` | `false` | Run interactive configuration wizard before starting |

If no config exists, an automatic setup runs with generated defaults and prints the admin password. Use `--wizard` for interactive setup.

```bash
modula serve                        # use ./modula.config.json
modula serve mysite                 # resolve from registry
modula serve mysite production      # specific environment
modula serve --wizard               # interactive setup first
```

### init

Initialize a ModulaCMS project in the current directory. Idempotent -- each step checks whether its output already exists and skips if so.

```bash
modula init
```

| Flag | Default | Description |
|------|---------|-------------|
| `--mode` | `interactive` | Init mode: `interactive`, `ci`, or `container` |
| `--admin-password` | | System admin password (required for `ci` mode) |
| `--name` | current directory name | Project name |

Steps performed:

1. Load or create the project registry (`~/.modula/configs.json`)
2. Create `modula/` project directory (if not present)
3. Write base config and environment overlays (local, dev, prod)
4. Register configs in the registry, set `local` as default environment
5. Generate localhost TLS certificates
6. Create and seed SQLite database (skipped for external databases or container mode)

```bash
modula init                                       # interactive
modula init --mode ci --admin-password s3cret!    # CI pipeline
modula init --mode container                      # Docker entrypoint
modula init --name my-site                        # custom project name
```

> **Good to know**: The `--yes` global flag maps to `--mode ci` for backward compatibility.

### status

Display the registration status, config files, environments, and available commands for the project in the current directory.

```bash
modula status
```

Shows:
- Base config path with `[ok]`/`[missing]` status
- Certificate directory and search index status
- All registered environment overlays with status markers
- Available `serve`, `connect`, and `tui` commands for each environment
- Troubleshooting tips when files are missing

If the current directory is not a registered project, suggests running `modula init`.

### tui

Launch the terminal UI without the server.

```bash
modula tui [project] [environment]
```

Connects directly to the database for local content management. No HTTP, HTTPS, or SSH servers are started.

```bash
modula tui                         # use ./modula.config.json
modula tui mysite                  # resolve from registry
modula tui mysite production       # specific environment
```

### connect

Launch the TUI for a registered project, or manage the project registry.

```bash
modula connect [project] [environment]
```

Resolves the config from the registry and launches the TUI via SSH or direct DB connection.

#### connect set

Register or update an environment for a project.

```bash
modula connect set <name> <environment> <config-path>
```

#### connect list

List registered projects and environments.

```bash
modula connect list
```

#### connect remove

Remove a project or a single environment.

```bash
modula connect remove <name>           # remove entire project
modula connect remove <name> --env dev # remove one environment
```

#### connect default

Set the default project or default environment.

```bash
modula connect default mysite              # set default project
modula connect default mysite production   # set default env for project
```

### config

Configuration management commands.

#### config show

Print the loaded configuration as JSON.

```bash
modula config show
modula config show --raw    # show overlay file only (requires --overlay)
```

#### config validate

Validate the configuration file.

```bash
modula config validate
modula config validate --json   # machine-readable output
```

JSON output format: `{"valid": true, "errors": []}`.

#### config set

Update a configuration field.

```bash
modula config set <key> <value>
modula config set db_password "newpass" --base   # write to base (requires --overlay)
```

#### config fields

List available configuration fields.

```bash
modula config fields                    # list all fields
modula config fields port               # show one field
modula config fields --category server  # filter by category
```

Categories: `server`, `database`, `storage`, `cors`, `cookie`, `oauth`, `observability`, `email`, `plugin`, `update`, `misc`.

#### config template

Print a complete config template with all fields and defaults.

```bash
modula config template
```

#### config overlay

Generate a minimal overlay config file for an environment.

```bash
modula config overlay --env staging
```

### db

Database management commands.

#### db init

Create database tables and seed bootstrap data.

```bash
modula db init
```

#### db wipe

Drop all database tables. Prompts for confirmation.

```bash
modula db wipe
```

#### db wipe-redeploy

Drop all tables, recreate schema, and re-seed with a new admin password.

```bash
modula db wipe-redeploy
```

#### db reset

Delete the database file (SQLite only).

```bash
modula db reset
```

#### db export

Dump the database to a SQL file.

```bash
modula db export
modula db export --file ./backup.sql   # custom output path
```

| Flag | Default | Description |
|------|---------|-------------|
| `--file` | auto-generated | Output file path |

### backup

Backup and restore commands.

#### backup create

Create a full backup (SQL dump + configured media paths).

```bash
modula backup create
```

#### backup restore

Restore from a backup archive. Prompts for confirmation.

```bash
modula backup restore <path>
```

#### backup list

List backup history from the database.

```bash
modula backup list
```

### deploy

Export, import, snapshot, push, and pull content data between environments.

All deploy subcommands support `--json` for machine-readable output.

#### deploy export

Export content data to a JSON file.

```bash
modula deploy export --file data.json
modula deploy export --file data.json --tables content_data,datatypes
modula deploy export --file data.json --include-plugins --json
```

#### deploy import

Import content data from a JSON export file.

```bash
modula deploy import data.json
modula deploy import data.json --dry-run
modula deploy import data.json --skip-backup
```

#### deploy push / pull

Sync data with a remote environment (configured in `deploy_environments`).

```bash
modula deploy push production
modula deploy pull staging --dry-run
```

#### deploy snapshot

Manage import snapshots.

```bash
modula deploy snapshot list
modula deploy snapshot show <id>
modula deploy snapshot restore <id>
```

#### deploy env

Manage deploy environments.

```bash
modula deploy env list
modula deploy env test production
```

### plugin

Plugin management commands. Some subcommands require a running server (marked "online").

| Flag | Default | Description |
|------|---------|-------------|
| `--token` | | Admin API token (overrides token file, for CI/CD) |

#### plugin list

List installed plugins.

```bash
modula plugin list
modula plugin list --json
```

#### plugin init

Create a new plugin scaffold.

```bash
modula plugin init <name>
modula plugin init my-plugin --version 1.0.0 --description "My plugin" --author "Name"
```

#### plugin validate

Validate a plugin without loading it.

```bash
modula plugin validate ./plugins/my-plugin
```

#### plugin info (online)

Show detailed plugin information from the running server.

```bash
modula plugin info <name>
modula plugin info <name> --json
```

#### plugin install

Install a discovered plugin (creates DB record).

```bash
modula plugin install <name>
modula plugin install <name> --yes
```

#### plugin reload (online)

Hot reload a plugin.

```bash
modula plugin reload <name>
```

#### plugin enable / disable (online)

Enable or disable a plugin.

```bash
modula plugin enable <name>
modula plugin disable <name>
```

#### plugin approve / revoke (online)

Approve or revoke plugin routes and hooks.

```bash
modula plugin approve <name> --all-routes
modula plugin approve <name> --all-hooks
modula plugin approve <name> --route "GET /tasks"
modula plugin approve <name> --hook "before_create:tasks"
modula plugin revoke <name> --all-routes
```

### pipeline

Pipeline management commands.

#### pipeline list

List all pipeline entries.

```bash
modula pipeline list
modula pipeline list --json
```

#### pipeline show

Show pipelines for a table, grouped by operation.

```bash
modula pipeline show <table>
modula pipeline show content_data --json
```

#### pipeline enable / disable / remove

Manage individual pipeline entries by ID.

```bash
modula pipeline enable <pipeline_id>
modula pipeline disable <pipeline_id>
modula pipeline remove <pipeline_id>
```

### cert

Certificate management.

#### cert generate

Generate self-signed SSL certificates for local development.

```bash
modula cert generate
```

### mcp

Start the MCP server over stdio (for AI tool integration).

```bash
modula mcp --url https://cms.example.com --api-key YOUR_KEY
```

Supports environment variables `MODULA_URL` and `MODULA_API_KEY`.

### version

Print version information and exit.

```bash
modula version
```

### update

Check for and apply binary updates from GitHub releases.

```bash
modula update
```

---

## Machine-Readable Output

Commands that support `--json` output structured JSON to stdout. Errors still go to stderr. This is designed for scripting, CI pipelines, and tool integration.

| Command | `--json` Support |
|---------|:---:|
| `config validate` | Yes |
| `plugin list` | Yes |
| `plugin info` | Yes |
| `pipeline list` | Yes |
| `pipeline show` | Yes |
| `deploy export` | Yes |
| `deploy import` | Yes |
| `deploy push` | Yes |
| `deploy pull` | Yes |
| `deploy snapshot list` | Yes |
| `deploy snapshot show` | Yes |
| `deploy snapshot restore` | Yes |
| `deploy env list` | Yes |
| `deploy env test` | Yes |

## Next Steps

- [Configuration](../getting-started/configuration.md) for all config fields
- [Plugin development](../extending/overview.md) for building plugins
- [Deployment](../deployment/production.md) for production setup
