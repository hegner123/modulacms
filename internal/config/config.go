package config

type ConfigOption string
type Endpoint string

type DbDriver string

const (
	OauthAuthURL  Endpoint = "oauth_auth_url"
	OauthTokenURL Endpoint = "oauth_token_url"
)

const (
	Sqlite DbDriver = "sqlite"
	Mysql  DbDriver = "mysql"
	Psql   DbDriver = "postgres"
)

type Config struct {
	Environment         string              `json:"environment"`
	Environment_Hosts   map[string]string   `json:"environment_hosts"`
	Port                string              `json:"port"`
	SSL_Port            string              `json:"ssl_port"`
	Cert_Dir            string              `json:"cert_dir"`
	Client_Site         string              `json:"client_site"`
	Admin_Site          string              `json:"admin_site"`
	SSH_Host            string              `json:"ssh_host"`
	SSH_Port            string              `json:"ssh_port"`
	Options             map[string][]any    `json:"options"`
	Log_Path            string              `json:"log_path"`
	Auth_Salt           string              `json:"auth_salt"`
	Cookie_Name         string              `json:"cookie_name"`
	Cookie_Duration     string              `json:"cookie_duration"`
	Db_Driver           DbDriver            `json:"db_driver"`
	Db_URL              string              `json:"db_url"`
	Db_Name             string              `json:"db_name"`
	Db_User             string              `json:"db_username"`
	Db_Password         string              `json:"db_password"`
	Bucket_Url          string              `json:"bucket_url"`
	Bucket_Region       string              `json:"bucket_region"`
	Bucket_Media        string              `json:"bucket_media"`
	Bucket_Backup       string              `json:"bucket_backup"`
	Bucket_Endpoint     string              `json:"bucket_endpoint"`
	Bucket_Access_Key   string              `json:"bucket_access_key"`
	Bucket_Secret_Key   string              `json:"bucket_secret_key"`
	Backup_Option       string              `json:"backup_option"`
	Backup_Paths        []string            `json:"backup_paths"`
	Oauth_Client_Id     string              `json:"oauth_client_id"`
	Oauth_Client_Secret string              `json:"oauth_client_secret"`
	Oauth_Scopes        []string            `json:"oauth_scopes"`
	Oauth_Endpoint      map[Endpoint]string `json:"oauth_endpoint"`
	Cors_Origins        []string            `json:"cors_origins"`
	Cors_Methods        []string            `json:"cors_methods"`
	Cors_Headers        []string            `json:"cors_headers"`
	Cors_Credentials    bool                `json:"cors_credentials"`
	Custom_Style_Path   string              `json:"custom_style_path"`
}

var DisableSystemTables ConfigOption = "disableSystemTables"
