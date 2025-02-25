package mTemplate

import (
	"testing"
)

func TestServeTemplate(t *testing.T) {
	test := []string{"../../templates/modula_base.html"}
	_, err := ParseTemplates(test)
	if err != nil {
		t.FailNow()
		return
	}
}
