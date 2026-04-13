package install_test

import (
	"fmt"
	"testing"

	"github.com/hegner123/modulacms/internal/install"
)

func TestNewInstallProgress(t *testing.T) {
	t.Parallel()

	p := install.NewInstallProgress()
	if p == nil {
		t.Fatal("NewInstallProgress() returned nil")
	}
}

func TestAddStep_Chaining(t *testing.T) {
	t.Parallel()

	p := install.NewInstallProgress()
	result := p.
		AddStep("step1", "first step", func() error { return nil }).
		AddStep("step2", "second step", func() error { return nil }).
		AddStep("step3", "third step", func() error { return nil })

	if result != p {
		t.Error("AddStep should return the same InstallProgress for chaining")
	}
}

func TestStep_Fields(t *testing.T) {
	t.Parallel()

	called := false
	action := func() error {
		called = true
		return nil
	}

	s := install.Step{
		Name:        "test step",
		Description: "a test step description",
		Action:      action,
	}

	if s.Name != "test step" {
		t.Errorf("Name = %q, want %q", s.Name, "test step")
	}
	if s.Description != "a test step description" {
		t.Errorf("Description = %q, want %q", s.Description, "a test step description")
	}

	err := s.Action()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("action was not called")
	}
}

func TestStep_ActionReturnsError(t *testing.T) {
	t.Parallel()

	expected := fmt.Errorf("step failed")
	s := install.Step{
		Name:        "failing step",
		Description: "this step fails",
		Action:      func() error { return expected },
	}

	err := s.Action()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != expected {
		t.Errorf("error = %v, want %v", err, expected)
	}
}
