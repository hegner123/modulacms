package permissions

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestPermissionEncoding(t *testing.T) {
	var res Role
	j, err := json.Marshal(AdminRole)
	if err != nil {
		t.Fatal(err)
	}
	js := string(j)

	fmt.Printf("%v\n", js)
	err = json.Unmarshal(j, &res)
    if err!=nil {
		t.Fatal(err)
    }
	fmt.Printf("\n%v\n",res)

}
