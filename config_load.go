package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

func loadConfig(verbose *bool) Config {
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatal("Error opening file:", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatal("Error reading file:", err)
	}

	var config Config
	if err := json.Unmarshal(bytes, &config); err != nil {
		log.Fatal("Error parsing JSON:", err)
	}
	if *verbose {
		fmt.Printf("%s\n", bytes)
	}
	return config
}
