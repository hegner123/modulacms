package config

import (
	"github.com/charmbracelet/huh"
	utility "github.com/hegner123/modulacms/internal/utility"
)

var (
	configForm *huh.Form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput(),
		),
	)
)

func CreateConfigCLI() {
	err := configForm.Run()
	if err != nil {
		utility.DefaultLogger.Fatal("", err)
	}

}
