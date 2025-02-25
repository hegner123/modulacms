package install

import "testing"

func TestConfigPathCheck(t *testing.T) {
	err := CheckConfigExists("")
	if err != nil {
		t.Fatal(err)
	}

}

func TestDbExists(t *testing.T){
    err := CheckDb("")
	if err != nil {
		t.Fatal(err)
	}

}
