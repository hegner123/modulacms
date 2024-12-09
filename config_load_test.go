package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func compareStructs(a, b interface{}) bool {
	if reflect.DeepEqual(a, b) {
		return true
	}

	// Get the type and value of the structs
	valA := reflect.ValueOf(a)
	valB := reflect.ValueOf(b)

	if valA.Type() != valB.Type() {
		fmt.Println("The structs are of different types.")
		return false
	}

	// Iterate through the fields of the struct
	for i := 0; i < valA.NumField(); i++ {
		fieldA := valA.Field(i)
		fieldB := valB.Field(i)
		fieldName := valA.Type().Field(i).Name

		// Compare field values
		if !reflect.DeepEqual(fieldA.Interface(), fieldB.Interface()) {
			fmt.Printf("Field '%s' is different: %v != %v\n", fieldName, fieldA.Interface(), fieldB.Interface())
			return false
		}
	}
	return false
}

func TestLoadConfig(t *testing.T) {
	fakeFlag := false
	conf := loadConfig(&fakeFlag)

	expected := Config{
		Port:                "8080",
		SSL_Port:            "8443",
		Client_Site:         "modulacms.com",
		Db_Driver:           "sqlite",
		Db_URL:              "./modula.db",
		Db_Name:             "modula.db",
		Db_Password:         "none",
		Bucket_Url:          "us-iad-10.linodeobjects.com",
		Bucket_Endpoint:     "backups.us-iad-10.linodeobjects.com",
		Bucket_Access_Key:   "RMK7Q10WV4AUMFAZYI7E",
		Bucket_Secret_Key:   "LNbFZDSi25erOCWdRbADU4hmeLw97W8IHHx20sk4",
		Backup_Option:       "",
		Backup_Paths:        []string{""},
		Oauth_Client_Id:     "Ov23liFoy8pVGnAnGgrE",
		Oauth_Client_Secret: "f57dda6a58faa59e4803f08efca11362478dcd3c",
		Oauth_Scopes:        []string{"profile", "profilePic"},
		Oauth_Endpoint:      map[Endpoint]string{oauthAuthURL: "https://github.com/login/oauth/authorize", oauthTokenURL: "https://github.com/login/oauth2/token"},
	}

	e := json.NewEncoder(os.Stdout)
	err := e.Encode(expected)
	if err != nil {
		logError("failed to encode ", err)
	}
	fmt.Println()

	res := compareStructs(conf, expected)
	if !res {
		t.FailNow()
	}
}
