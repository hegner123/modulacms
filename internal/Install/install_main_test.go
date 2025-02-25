package install

import "testing"

func TestMainInstall(t *testing.T) {
	verbose := false
	err := InstallMain("", &verbose)
	if err != nil {
		t.Fatal(err)
	}
}
