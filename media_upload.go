package main

import (
	"bytes"
	"fmt"
	"os"
)

func handleMediaUpload(file *bytes.Buffer, fPath string) {
	newPath := fmt.Sprintf("./tmp-%s", fPath)

	err := os.WriteFile(newPath, file.Bytes(), os.FileMode(0555))
	if err != nil {
		fmt.Printf("%s\n", err)
	}
    f,err:= os.Open(newPath)
    if err!=nil {
        fmt.Printf("%s\n",err)
    }

	optimized := optimizeUpload(f, fPath)
	if optimized == 1 {
		fmt.Printf("Couldn't Optimize file\n")
	}

}
