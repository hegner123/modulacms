package install

import (
	"os"
	"testing"
)

func TestCreateDefaultConfig(t *testing.T){
    err:= CreateDefaultConfig("test-config.json")
    if err!=nil {
        t.Fatal(err)
    }
    os.Remove("test-config.json")

}
