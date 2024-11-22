package main

type Database struct {
	DB string
}

type Config struct {
	Port            string   `json:"port"`
	SSLPort         string   `json:"ssl_port"`
	ClientSite      string   `json:"client_site"`
	DB_DRIVER       string   `json:"db_driver"`
	DB_URL          string   `json:"db_url"`
	DB_NAME         string   `json:"db_name"`
	DB_PASSWORD     string   `json:"db_password"`
	Bucket_URL      string   `json:"bucket_url"`
	Bucket_PASSWORD string   `json:"bucket_password"`
	Backup_Option   string   `json:"backup_option"`
	Backup_Paths    []string `json:"backup_Path"`
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
