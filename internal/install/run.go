package install

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/utility"
)


func RunInstall(v *bool) {

	dir, err := os.Getwd()
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}

	s := fmt.Sprintf("Would you like to install ModulaCMS at \n %s\n", dir)
	runInstall := false
	c := huh.NewConfirm().Title(s).Value(&runInstall)
	err = c.Run()
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}

	if !runInstall {
		os.Exit(0)
	}

	iarg, err := RunInstallIO()
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}
	var f *os.File

	_, err = os.Stat(iarg.ConfigPath)

	if err == nil {

		f, err = os.Open(iarg.ConfigPath)
		if err != nil {
			utility.DefaultLogger.Error("", err)
		}

	} else {
		f, err = os.Create(iarg.ConfigPath)
		if err != nil {
			utility.DefaultLogger.Error("", err)
		}
	}
	j, err := utility.FormatJSON(iarg.Config)
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}
	_, err = fmt.Fprintln(f, j)
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}
	_, err = CheckInstall(&iarg.Config,v)
	if err != nil {
		RunInstall(v)
	}

}
