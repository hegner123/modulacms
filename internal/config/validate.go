package config

import "fmt"

// ValidationResult holds the outcome of a configuration validation.
type ValidationResult struct {
	Valid           bool
	Errors          []string
	Warnings        []string
	RestartRequired []string
}

// Validate checks a Config for required fields and valid values.
func Validate(c Config) ValidationResult {
	var result ValidationResult

	if c.Db_Driver == "" {
		result.Errors = append(result.Errors, "db_driver is required")
	} else {
		switch c.Db_Driver {
		case Sqlite, Mysql, Psql:
			// valid
		default:
			result.Errors = append(result.Errors, fmt.Sprintf("db_driver %q is not valid (sqlite, mysql, postgres)", c.Db_Driver))
		}
	}

	if c.Db_URL == "" {
		result.Errors = append(result.Errors, "db_url is required")
	}

	if c.Port == "" {
		result.Errors = append(result.Errors, "port is required")
	}

	if c.SSH_Port == "" {
		result.Errors = append(result.Errors, "ssh_port is required")
	}

	if c.Output_Format != "" && !IsValidOutputFormat(string(c.Output_Format)) {
		result.Errors = append(result.Errors, fmt.Sprintf("output_format %q is not valid", c.Output_Format))
	}

	if c.Observability_Sample_Rate < 0 || c.Observability_Sample_Rate > 1 {
		result.Warnings = append(result.Warnings, "observability_sample_rate should be between 0.0 and 1.0")
	}

	if c.Observability_Traces_Rate < 0 || c.Observability_Traces_Rate > 1 {
		result.Warnings = append(result.Warnings, "observability_traces_rate should be between 0.0 and 1.0")
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// ValidateUpdate compares the current and proposed configs and returns
// validation results including which changed fields require a restart.
func ValidateUpdate(current, proposed Config) ValidationResult {
	result := Validate(proposed)

	// Check each non-hot-reloadable field for changes.
	for _, f := range FieldRegistry {
		if f.HotReloadable {
			continue
		}
		currentVal := configFieldString(current, f.JSONKey)
		proposedVal := configFieldString(proposed, f.JSONKey)
		if currentVal != proposedVal {
			result.RestartRequired = append(result.RestartRequired,
				fmt.Sprintf("%s changed from %q to %q", f.JSONKey, currentVal, proposedVal))
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Field %q requires server restart to take effect", f.JSONKey))
		}
	}

	return result
}

// configFieldString returns the string representation of a config field
// by its JSON key. Uses explicit field mapping (no reflection).
func configFieldString(c Config, key string) string {
	switch key {
	case "environment":
		return c.Environment
	case "os":
		return c.OS
	case "port":
		return c.Port
	case "ssl_port":
		return c.SSL_Port
	case "cert_dir":
		return c.Cert_Dir
	case "ssh_host":
		return c.SSH_Host
	case "ssh_port":
		return c.SSH_Port
	case "client_site":
		return c.Client_Site
	case "admin_site":
		return c.Admin_Site
	case "log_path":
		return c.Log_Path
	case "auth_salt":
		return c.Auth_Salt
	case "node_id":
		return c.Node_ID
	case "space_id":
		return c.Space_ID
	case "output_format":
		return string(c.Output_Format)
	case "custom_style_path":
		return c.Custom_Style_Path
	case "max_upload_size":
		return fmt.Sprintf("%d", c.Max_Upload_Size)
	case "db_driver":
		return string(c.Db_Driver)
	case "db_url":
		return c.Db_URL
	case "db_name":
		return c.Db_Name
	case "db_username":
		return c.Db_User
	case "db_password":
		return c.Db_Password
	case "bucket_region":
		return c.Bucket_Region
	case "bucket_media":
		return c.Bucket_Media
	case "bucket_backup":
		return c.Bucket_Backup
	case "bucket_endpoint":
		return c.Bucket_Endpoint
	case "bucket_access_key":
		return c.Bucket_Access_Key
	case "bucket_secret_key":
		return c.Bucket_Secret_Key
	case "bucket_public_url":
		return c.Bucket_Public_URL
	case "bucket_default_acl":
		return c.Bucket_Default_ACL
	case "bucket_force_path_style":
		return fmt.Sprintf("%t", c.Bucket_Force_Path_Style)
	case "backup_option":
		return c.Backup_Option
	case "cors_credentials":
		return fmt.Sprintf("%t", c.Cors_Credentials)
	case "cookie_name":
		return c.Cookie_Name
	case "cookie_duration":
		return c.Cookie_Duration
	case "cookie_secure":
		return fmt.Sprintf("%t", c.Cookie_Secure)
	case "cookie_samesite":
		return c.Cookie_SameSite
	case "oauth_client_id":
		return c.Oauth_Client_Id
	case "oauth_client_secret":
		return c.Oauth_Client_Secret
	case "oauth_provider_name":
		return c.Oauth_Provider_Name
	case "oauth_redirect_url":
		return c.Oauth_Redirect_URL
	case "oauth_success_redirect":
		return c.Oauth_Success_Redirect
	case "observability_enabled":
		return fmt.Sprintf("%t", c.Observability_Enabled)
	case "observability_provider":
		return c.Observability_Provider
	case "observability_dsn":
		return c.Observability_DSN
	case "observability_environment":
		return c.Observability_Environment
	case "observability_release":
		return c.Observability_Release
	case "observability_sample_rate":
		return fmt.Sprintf("%g", c.Observability_Sample_Rate)
	case "observability_traces_rate":
		return fmt.Sprintf("%g", c.Observability_Traces_Rate)
	case "observability_send_pii":
		return fmt.Sprintf("%t", c.Observability_Send_PII)
	case "observability_debug":
		return fmt.Sprintf("%t", c.Observability_Debug)
	case "observability_server_name":
		return c.Observability_Server_Name
	case "observability_flush_interval":
		return c.Observability_Flush_Interval
	case "plugin_enabled":
		return fmt.Sprintf("%t", c.Plugin_Enabled)
	case "plugin_directory":
		return c.Plugin_Directory
	case "plugin_max_vms":
		return fmt.Sprintf("%d", c.Plugin_Max_VMs)
	case "plugin_timeout":
		return fmt.Sprintf("%d", c.Plugin_Timeout)
	case "plugin_max_ops":
		return fmt.Sprintf("%d", c.Plugin_Max_Ops)
	case "plugin_hot_reload":
		return fmt.Sprintf("%t", c.Plugin_Hot_Reload)
	case "plugin_max_failures":
		return fmt.Sprintf("%d", c.Plugin_Max_Failures)
	case "plugin_reset_interval":
		return c.Plugin_Reset_Interval
	case "plugin_rate_limit":
		return fmt.Sprintf("%d", c.Plugin_Rate_Limit)
	case "plugin_max_routes":
		return fmt.Sprintf("%d", c.Plugin_Max_Routes)
	case "plugin_max_request_body":
		return fmt.Sprintf("%d", c.Plugin_Max_Request_Body)
	case "plugin_max_response_body":
		return fmt.Sprintf("%d", c.Plugin_Max_Response_Body)
	case "update_auto_enabled":
		return fmt.Sprintf("%t", c.Update_Auto_Enabled)
	case "update_check_interval":
		return c.Update_Check_Interval
	case "update_channel":
		return c.Update_Channel
	case "update_notify_only":
		return fmt.Sprintf("%t", c.Update_Notify_Only)
	default:
		return ""
	}
}

// ConfigFieldString is the exported accessor for reading a config field
// value by its JSON key name. Returns an empty string for unknown keys.
func ConfigFieldString(c Config, key string) string {
	return configFieldString(c, key)
}
