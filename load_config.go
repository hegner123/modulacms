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
		fmt.Printf(`
            DB URL:%s, 
            DB NAME: %s, 
            DB Password: %s,
            Bucket URL: %s, 
            Bucket Password: %s
            `, config.DB_URL, config.DB_NAME,
			config.DB_PASSWORD, config.Bucket_URL, config.Bucket_PASSWORD)

	}
	return config

}
