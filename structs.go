package main

type Database struct {
	DB string
}

type Config struct {
	Port                string              `json:"port"`
	SSL_Port            string              `json:"ssl_port"`
	Client_Site         string              `json:"client_site"`
	Db_Driver           string              `json:"db_driver"`
	Db_URL              string              `json:"db_url"`
	Db_Name             string              `json:"db_name"`
	Db_Password         string              `json:"db_password"`
	Bucket_Url          string              `json:"bucket_url"`
	Bucket_Password     string              `json:"bucket_password"`
	Backup_Option       string              `json:"backup_option"`
	Backup_Paths        []string            `json:"backup_path"`
	Oauth_Client_Id     string              `json:"oauth_client_id"`
	Oauth_Client_Secret string              `json:"oauth_client_secret"`
	Oauth_Scopes        []string            `json:"oauth_scopes"`
	Oauth_Endpoint      map[Endpoint]string `json:"oauth_endpoint"`
}

type FieldType struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Example     interface{} `json:"example"`
}

type Schema struct {
	ObjectName string      `json:"objectName"`
	Fields     []FieldType `json:"fields"`
}

type Backup struct {
	Hash    string
	DbFile  string
	Archive string
}
