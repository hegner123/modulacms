package main

type PluginAPI struct {
	ID         int                 `json:"id"`
	Name       string              `json:"name"`
	Version    string              `json:"version"`
	Author     string              `json:"author"`
	AuthorUrl  string              `json:"authorUrl"`
	AddsTables bool                `json:"addsTables"`
	Tables     []PluginAPITable    `json:"tables"`
	Relations  []PluginAPIRelation `json:"relations"`
}

type PluginAPITable struct {
	ID          int                 `json:"id"`
	Name        string              `json:"name"`
	Columns     []PluginAPIColumn   `json:"columns"`
	PrimaryKey  int                 `json:"primaryKey"`
	ForeignKeys []PluginAPIRelation `json:"ForeignKeys"`
}

type PluginAPIColumn struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"defaultValue"`
}

type PluginAPIRelation struct {
	ID              int `json:"id"`
	TableID         int `json:"tableId"`
	ColumnID        int `json:"columnId"`
	ForeignTableId  int `json:"foreignTableId"`
	ForeignColumnID int `json:"foreignColumnId"`
}
