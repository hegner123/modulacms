package main

import (
	"fmt"
	"time"
)

func handleCompletedMediaUpload(tmpFile string, fName string) {
	now := time.Now()
	year := now.Year()
	month := now.Month()
	newPath := fmt.Sprintf("./media/%d/%d/%s", year, month, fName)

	optimized := optimizeUpload(tmpFile, fName)

    //TODO: write paths to optimized files to db

	for _, v := range optimized {
		err := objectUpload(v, newPath)
		if err != nil {
			logError("failed to upload optimized files to blob storage ", err)
		}

	}

}
