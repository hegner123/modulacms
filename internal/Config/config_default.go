package config

import "encoding/json"

func DefaultConfig() Config {
	var c Config
	c.Port = "8080"
	c.SSL_Port = "8443"
	c.Client_Site = "localhost"
	c.Admin_Site = "localhost"
	c.Db_Driver = "sqlite"
	c.Db_Name = "modula.db"
	c.Db_URL = "./modula.db"
	c.Db_Password = ""
	c.Backup_Option = ""
	c.Backup_Paths = []string{""}
	
	// Default CORS settings
	c.Cors_Origins = []string{"http://localhost:3000"}
	c.Cors_Methods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	c.Cors_Headers = []string{"Content-Type", "Authorization"}
	c.Cors_Credentials = true
	
	return c
}

func (c Config) JSON() []byte {
	j, err := json.Marshal(c)
	if err != nil {
		return []byte{}
	}
	return j
}
