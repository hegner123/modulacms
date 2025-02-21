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
	return c
}

func (c Config) JSON() []byte {
	j, err := json.Marshal(c)
	if err != nil {
		return []byte{}
	}
	return j
}
