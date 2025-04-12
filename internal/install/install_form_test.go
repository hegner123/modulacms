package install

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestInstallArg(t *testing.T) {
	args, err := RunInstallIO()
	if err != nil {
		t.Fatal(err)
	}
	j, err := json.Marshal(args)
    if err!=nil {
		t.Fatal(err)
    }
	fmt.Println(string(j))

}
