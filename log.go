package main

import (
	"fmt"
	"io"
	"os"
	"strings"
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

func popError(err error) string {
	unwrappedErr := strings.Split(err.Error(), " ")
	msg := fmt.Sprint(unwrappedErr[len(unwrappedErr)-1])
	return msg
}

func logError(message string, err error, args ...any) {
	var messageParts []string
	messageParts = append(messageParts, message)
	for _, arg := range args {
		switch v := arg.(type) {
		case fmt.Stringer: // If the type implements fmt.Stringer, use String()
			messageParts = append(messageParts, v.String())
		default:
			messageParts = append(messageParts, fmt.Sprintf("%+v", arg)) // Format structs nicely
		}
	}
	fullMessage := strings.Join(messageParts, " ")

	// Format the final error message
	er := fmt.Errorf("%sErr: %s\n%v\n%s", RED, fullMessage, err, RESET)
	if er != nil {
		fmt.Printf("%s\n", er)
	}
}

func pLog(args ...any) {
	fmt.Printf("%s", BLUE)
	for _, arg := range args {
		fmt.Print(arg)
	}
	fmt.Printf("%s\n", RESET)
}
