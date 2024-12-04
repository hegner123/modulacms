package main

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

type S3Credintials struct {
	AccessKey string
	SecretKey string
	URL       string
}
