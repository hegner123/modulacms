package main

import (
	"fmt"
	"os"
	"testing"

	install "github.com/hegner123/modulacms/internal/Install"
)

type GlobalTestingState struct {
	Initialized bool
}

var GlobalTesting GlobalTestingState

func setup() {
	fmt.Printf("TestMain setup\n")
	GlobalTesting.Initialized = true
}

func teardown() {
	fmt.Printf("TestMain teardown\n")
	GlobalTesting.Initialized = false
}

func TestMain(m *testing.M) {
	fmt.Printf("TestMain init\n")
	GlobalTesting.Initialized = false
	setup()
	code := m.Run()
	teardown()
	fmt.Printf("TestMain exit\n")
	os.Exit(code)
}

func TestInit(t *testing.T) {
	s,err := install.CheckInstall()
    if err!=nil {
		t.FailNow()
    }

	if !s.UseSSL || !s.DbFileExists {
		t.FailNow()
	}

}
