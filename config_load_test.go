package main

import (
	"fmt"
	"reflect"
	"testing"
)
func compareStructs(a, b interface{})bool {
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
        Port:"8080",
        SSL_Port: "443",
        Client_Site: "example.com",
        Db_Driver: "sqlite",
        Db_URL: "default",
        Db_Name: "default",
        Db_Password:"none",
        Bucket_Url: "local",
        Bucket_Password:"none",
        Backup_Option:"",
        Backup_Paths: []string{},
        Oauth_Client_Id:"clientId",
        Oauth_Client_Secret: "clientSecret",
        Oauth_Scopes:[]string{"profile","profilePic"},
        Oauth_Endpoint: map[Endpoint]string{oauthAuthURL:"https://provider.com/o/oauth2/auth", oauthTokenURL:"https://provider.com/o/oauth2/token"}, 
    }
    
    res:=compareStructs(conf,expected)
    if !res{
        t.FailNow()
    }


}
