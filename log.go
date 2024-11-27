package main

import (
	"fmt"
	"io"
	"os"
)

func logGetVersion() string {
	file, err := os.Open("version.json")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return "Error reading file:"
	}
	return string(bytes)
}

func logError(message string, err error) {
	fmt.Print("err\n")
	er := fmt.Errorf("%s %w\n", message, err)
	if er != nil {
		fmt.Printf("%s\n", er)
	}
}
