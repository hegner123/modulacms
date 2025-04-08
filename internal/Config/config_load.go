package config

import (
	"encoding/json"
	"io"
	"os"

	utility "github.com/hegner123/modulacms/internal/utility"
)

var file *os.File
var err error
var config Config
var Env Config

// TODO Add error handling for when modula doesn't have permissions
// to wirte to error log path

func LoadConfig(verbose *bool, altConfig string) Config {
	if altConfig != "" {
		file, err = os.Open(altConfig)
		if *verbose {
			utility.DefaultLogger.Info("load alt config", altConfig)
		}
		if err != nil {
			utility.DefaultLogger.Error("Error opening file:", err)
		}
	} else {
		file, err = os.Open("config.json")
		if err != nil {
			utility.DefaultLogger.Fatal("Error opening file:", err)
		}
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		utility.DefaultLogger.Fatal("Error reading file:", err)
	}

	if err := json.Unmarshal(bytes, &config); err != nil {
		utility.DefaultLogger.Fatal("Error parsing JSON:", err)
	}
	if *verbose {
		utility.DefaultLogger.Finfo("", string(bytes))
	}
	Env = config
	return config
}
