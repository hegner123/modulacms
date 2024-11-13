package main

import (
	"bytes"
	"fmt"
	"os"
	"time"
)

func handleMediaUpload(file *bytes.Buffer, fName string) {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	newPath := fmt.Sprintf("./media/%s/%s/%s", year, month, fName)

	err := os.WriteFile(newPath, file.Bytes(), os.FileMode(0555))
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	f, err := os.Open(newPath)
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	optimized := optimizeUpload(f, fName)
	if optimized == 1 {
		fmt.Printf("Couldn't Optimize file\n")
	} else {
		fmt.Printf("Optimized file\n")
	}

}
