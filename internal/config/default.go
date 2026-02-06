package config

import (
	"encoding/base64"
	"encoding/json"
	"runtime"
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
	hosts["http-only"] = "localhost"

	var c Config
	c.Environment = "development"
	c.OS = runtime.GOOS
	c.Environment_Hosts = hosts
	c.Port = ":8080"
	c.SSL_Port = ":4000"
    c.Cert_Dir = "./"
	c.Client_Site = "localhost"
	c.Admin_Site = "localhost"
	c.SSH_Host = "localhost"
	c.SSH_Port = "2233"
	c.Log_Path = "./"
	c.Auth_Salt = enc
    c.Cookie_Name = "modula_cms"
    c.Cookie_Duration = "1w"
	c.Cookie_Secure = false      // Set to true in production with HTTPS
	c.Cookie_SameSite = "lax"    // Options: "strict", "lax", "none"
	c.Db_Driver = "sqlite"
	c.Db_Name = "modula.db"
	c.Db_URL = "./modula.db"
	c.Db_Password = ""
	c.Backup_Option = "./"
	c.Backup_Paths = []string{""}
	c.Bucket_Force_Path_Style = true
	c.Bucket_Region = "us-east-1"

	// Default CORS settings
	c.Cors_Origins = []string{"http://localhost:3000"}
	c.Cors_Methods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	c.Cors_Headers = []string{"Content-Type", "Authorization"}
	c.Cors_Credentials = true

	// Default update settings
	c.Update_Auto_Enabled = false
	c.Update_Check_Interval = "startup"
	c.Update_Channel = "stable"
	c.Update_Notify_Only = false

	// Default OAuth settings (configure per provider)
	c.Oauth_Client_Id = ""
	c.Oauth_Client_Secret = ""
	c.Oauth_Scopes = []string{"openid", "profile", "email"}
	c.Oauth_Endpoint = map[Endpoint]string{
		OauthAuthURL:     "",
		OauthTokenURL:    "",
		OauthUserInfoURL: "",
	}
	c.Oauth_Provider_Name = ""
	c.Oauth_Redirect_URL = ""
	c.Oauth_Success_Redirect = "/"

	// Default observability settings
	c.Observability_Enabled = false
	c.Observability_Provider = "console"
	c.Observability_DSN = ""
	c.Observability_Environment = "development"
	c.Observability_Release = ""
	c.Observability_Sample_Rate = 1.0
	c.Observability_Traces_Rate = 0.1
	c.Observability_Send_PII = false
	c.Observability_Debug = false
	c.Observability_Server_Name = ""
	c.Observability_Flush_Interval = "30s"
	c.Observability_Tags = map[string]string{}

	c.KeyBindings = DefaultKeyMap()

	return c
}

func (c Config) JSON() []byte {
	j, err := json.Marshal(c)
	if err != nil {
		return []byte{}
	}
	return j
}
