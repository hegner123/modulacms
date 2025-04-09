package config

import (
	"encoding/json"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/utility"
)

type TestColor struct {
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Padding   []int
}

func TestMarshalCompleteConfig(t *testing.T) {
	eh := map[string]string{}
	eh["local"] = "localhost"
	eh["development"] = "dev.host.com"
	eh["staging"] = "stage.host.com"
	eh["prod"] = "host.com"
	options := map[string][]any{}
	res := make([]any, 0)
	tc := TestColor{
		Primary:   lipgloss.Color("#FFFFFF"),
		Secondary: lipgloss.Color("#C0C0C0"),
		Padding:   []int{1, 1, 1, 1},
	}
	res = append(res, tc)
	options["option1"] = res
	endpoints := map[Endpoint]string{}
	endpoints[OauthAuthURL] = "OauthAuthURL"
	endpoints[OauthTokenURL] = "OauthAuthURL"
	c := Config{
		Environment:         "local",
		Environment_Hosts:   eh,
		Port:                "4000",
		SSL_Port:            "4000",
		Cert_Dir:            "/Users/Demo/Path/certs",
		Client_Site:         "example.com",
		Admin_Site:          "admin_example.com",
		SSH_Host:            "admin_example.com",
		SSH_Port:            "2222",
		Options:             options,
		Log_Path:            "/var/log/modulacms.log",
		Auth_Salt:           "CustomSalt",
		Cookie_Name:         "My_Site",
		Cookie_Duration:     "1m2w7d3h30m20s",
		Db_Driver:           "sqlite | mysql | postgres",
		Db_URL:              "<file_path>|<url>",
		Db_Name:             "Db_Name",
		Db_User:             "Db_Name_User",
		Db_Password:         "Db_Password",
		Bucket_Url:          "https://region.provider.misc",
		Bucket_Region:       "S3_REGION",
		Bucket_Media:        "/project/media_folder",
		Bucket_Backup:       "/project/backup_folder",
		Backup_Option:       "options_placeholder",
		Backup_Paths:        []string{"/development", "/staging", "/prod", "/deploy"},
		Oauth_Client_Id:     "<Oauth_Client_Id>",
		Oauth_Client_Secret: "< Oauth_Client_Secret >",
		Oauth_Scopes:        []string{"<Oauth_CLIENT_SCOPE>"},
		Oauth_Endpoint:      endpoints,
		Cors_Methods:        []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		Cors_Headers:        []string{"Content-Type", "Authorization"},
		Cors_Credentials:    true,
	}
	jc, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}

	utility.DefaultLogger.Fblank(string(jc))

}
