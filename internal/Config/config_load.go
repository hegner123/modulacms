package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

var file *os.File
var err error
var config Config
var Env Config

func LoadConfig(verbose *bool, altConfig string) Config {
	if altConfig != "" {
		file, err = os.Open(altConfig)
		if *verbose {
			fmt.Println("load alt config")
			fmt.Println(altConfig)
		}
		if err != nil {
			log.Fatal("Error opening file:", err)
		}
	} else {
		file, err = os.Open("config.json")
		if err != nil {
			log.Fatal("Error opening file:", err)
		}
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatal("Error reading file:", err)
	}

	if err := json.Unmarshal(bytes, &config); err != nil {
		log.Fatal("Error parsing JSON:", err)
	}
	if *verbose {
		fmt.Printf("%s\n", bytes)
	}
	Env = config
	return config
}
