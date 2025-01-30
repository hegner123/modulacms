package main

import (
	"fmt"
	"os"
	"testing"

)

type GlobalTestingState struct {
	Initialized bool
}

var globalTestingState GlobalTestingState

func setup() {
	fmt.Printf("TestMain setup\n")
	globalTestingState.Initialized = true
}

func teardown() {
	fmt.Printf("TestMain teardown\n")
	globalTestingState.Initialized = false
}

func TestMain(m *testing.M) {
	fmt.Printf("TestMain init\n")
	globalTestingState.Initialized = false
	setup()
	code := m.Run()
	teardown()
	fmt.Printf("TestMain exit\n")
	os.Exit(code)
}

