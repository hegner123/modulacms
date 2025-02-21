package config

type ConfigOption string
type Endpoint string

const (
	oauthAuthURL  Endpoint = "oauth_auth_url"
	oauthTokenURL Endpoint = "oauth_token_url"
)

type Config struct {
	Port                string              `json:"port"`
	SSL_Port            string              `json:"ssl_port"`
	Client_Site         string              `json:"client_site"`
	Admin_Site          string              `json:"admin_site"`
	Options             map[string][]any    `json:"options"`
	Db_Driver           string              `json:"db_driver"`
	Db_URL              string              `json:"db_url"`
	Db_Name             string              `json:"db_name"`
	Db_Password         string              `json:"db_password"`
	Bucket_Url          string              `json:"bucket_url"`
	Bucket_Endpoint     string              `json:"bucket_endpoint"`
	Bucket_Access_Key   string              `json:"bucket_access_key"`
	Bucket_Secret_Key   string              `json:"bucket_secret_key"`
	Backup_Option       string              `json:"backup_option"`
	Backup_Paths        []string            `json:"backup_paths"`
	Oauth_Client_Id     string              `json:"oauth_client_id"`
	Oauth_Client_Secret string              `json:"oauth_client_secret"`
	Oauth_Scopes        []string            `json:"oauth_scopes"`
	Oauth_Endpoint      map[Endpoint]string `json:"oauth_endpoint"`
}

var DisableSystemTables ConfigOption = "disableSystemTables"
