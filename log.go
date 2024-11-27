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
	er := fmt.Errorf("%serr\n %s\n\n %v\n %s", RED, message, err, RESET)
	if er != nil {
		fmt.Printf("%s\n", er)
	}
}

func pLog(args ...any) {
	fmt.Printf("%s", BLUE)
	for _, arg := range args {
		fmt.Print( arg)
	}
	fmt.Printf("%s\n", RESET)
}
