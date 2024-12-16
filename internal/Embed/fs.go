package mEmbed

import (
	"fmt"
	"os"
)

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.Mode().IsRegular()
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func ReadEmbedFS(directory string) {
	dir, err := sqlFiles.ReadDir(directory)
	if err != nil {
		logError("error in ReadEmbedFS ", err)
	}
	for key, value := range dir {
		fmt.Printf("%d:%s\n", key, value)
	}
}
