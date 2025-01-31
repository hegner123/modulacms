package utility

import (
	"fmt"
	"testing"
)

func TestLog(t *testing.T) {
	_, err := GetVersion()
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
}
