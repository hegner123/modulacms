// install-test-harness exposes individual install form functions as subcommands
// for interactive testing via tui-test-ghost. Each subcommand runs one Get*
// function and prints only the fields that subcommand sets as JSON on stdout.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hegner123/modulacms/internal/install"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: install-test-harness <subcommand>")
		fmt.Fprintln(os.Stderr, "subcommands: use-default, config-path, ports, environments,")
		fmt.Fprintln(os.Stderr, "  db-driver, sqlite-setup, fullsql-setup, domains, cors,")
		fmt.Fprintln(os.Stderr, "  cert-dir, cookie, output-format, oauth-optional,")
		fmt.Fprintln(os.Stderr, "  admin-password, buckets")
		os.Exit(1)
	}

	iarg := &install.InstallArguments{}
	var err error
	var result any

	switch os.Args[1] {
	case "use-default":
		err = install.GetUseDefault(iarg)
		result = map[string]any{"use_default_config": iarg.UseDefaultConfig}

	case "config-path":
		err = install.GetConfigPath(iarg)
		result = map[string]any{"config_path": iarg.ConfigPath}

	case "ports":
		err = install.GetPorts(iarg)
		result = map[string]any{
			"port":     iarg.Config.Port,
			"ssl_port": iarg.Config.SSL_Port,
			"ssh_port": iarg.Config.SSH_Port,
		}

	case "environments":
		err = install.GetEnvironments(iarg)
		result = map[string]any{"environment_hosts": iarg.Config.Environment_Hosts}

	case "db-driver":
		err = install.GetDbDriver(iarg)
		result = map[string]any{
			"db_driver":        iarg.DB_Driver,
			"config_db_driver": iarg.Config.Db_Driver,
		}

	case "sqlite-setup":
		err = install.GetLiteSqlSetup(iarg)
		result = map[string]any{
			"db_url":  iarg.Config.Db_URL,
			"db_name": iarg.Config.Db_Name,
		}

	case "fullsql-setup":
		err = install.GetFullSqlSetup(iarg)
		result = map[string]any{
			"db_url":      iarg.Config.Db_URL,
			"db_name":     iarg.Config.Db_Name,
			"db_user":     iarg.Config.Db_User,
			"db_password": iarg.Config.Db_Password,
		}

	case "domains":
		err = install.GetDomains(iarg)
		result = map[string]any{
			"client_site": iarg.Config.Client_Site,
			"admin_site":  iarg.Config.Admin_Site,
		}

	case "cors":
		err = install.GetCORS(iarg)
		result = map[string]any{
			"cors_origins":     iarg.Config.Cors_Origins,
			"cors_credentials": iarg.Config.Cors_Credentials,
			"cors_methods":     iarg.Config.Cors_Methods,
			"cors_headers":     iarg.Config.Cors_Headers,
		}

	case "cert-dir":
		err = install.GetCertDir(iarg)
		result = map[string]any{"cert_dir": iarg.Config.Cert_Dir}

	case "cookie":
		err = install.GetCookie(iarg)
		result = map[string]any{"cookie_name": iarg.Config.Cookie_Name}

	case "output-format":
		err = install.GetOutputFormat(iarg)
		result = map[string]any{"output_format": iarg.Config.Output_Format}

	case "oauth-optional":
		err = install.GetOAuthOptional(iarg)
		result = map[string]any{
			"oauth_client_id":     iarg.Config.Oauth_Client_Id,
			"oauth_client_secret": iarg.Config.Oauth_Client_Secret,
			"oauth_endpoint":      iarg.Config.Oauth_Endpoint,
			"oauth_provider_name": iarg.Config.Oauth_Provider_Name,
		}

	case "admin-password":
		err = install.GetAdminPassword(iarg)
		result = map[string]any{
			"has_hash": iarg.AdminPasswordHash != "",
			"hash_len": len(iarg.AdminPasswordHash),
		}

	case "buckets":
		err = install.GetBuckets(iarg)
		result = map[string]any{
			"bucket_access_key":      iarg.Config.Bucket_Access_Key,
			"bucket_secret_key":      iarg.Config.Bucket_Secret_Key,
			"bucket_region":          iarg.Config.Bucket_Region,
			"bucket_endpoint":        iarg.Config.Bucket_Endpoint,
			"bucket_media":           iarg.Config.Bucket_Media,
			"bucket_backup":          iarg.Config.Bucket_Backup,
			"bucket_force_path_style": iarg.Config.Bucket_Force_Path_Style,
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", os.Args[1])
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	out, marshalErr := json.MarshalIndent(result, "", "  ")
	if marshalErr != nil {
		fmt.Fprintf(os.Stderr, "marshal error: %v\n", marshalErr)
		os.Exit(1)
	}
	fmt.Println(string(out))
}
