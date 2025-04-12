package config

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"time"
)

func DefaultConfig() Config {
	salt := time.Now().Unix()
	saltString := strconv.FormatInt(salt, 10)
	enc := base64.RawStdEncoding.EncodeToString([]byte(saltString))
	hosts := map[string]string{}
	hosts["local"] = "localhost"
	hosts["development"] = "localhost"
	hosts["staging"] = "localhost"
	hosts["production"] = "localhost"

	var c Config
	c.Environment = "development"
	c.Environment_Hosts = hosts
	c.Port = "8080"
	c.SSL_Port = "4000"
    c.Cert_Dir = "./"
	c.Client_Site = "localhost"
	c.Admin_Site = "localhost"
	c.SSH_Host = "localhost"
	c.SSH_Port = "2233"
	c.Log_Path = "./"
	c.Auth_Salt = enc
    c.Cookie_Name = "modula_cms"
    c.Cookie_Duration = "1w"
	c.Db_Driver = "sqlite"
	c.Db_Name = "modula.db"
	c.Db_URL = "./modula.db"
	c.Db_Password = ""
	c.Backup_Option = "./"
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
