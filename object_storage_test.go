package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"testing"
)

func TestObjectStorage(t *testing.T) {
	config := Config{}

	fmt.Println("Load config from file")

	file, err := os.Open("testing-config.json")
	if err != nil {
		logError("failed to open config ", err)
		_, file, line, _ := runtime.Caller(0)
		fmt.Printf("Current line number: %s:%d\n", file, line)
		t.FailNow()
	}
	defer file.Close()
	c, err := io.ReadAll(file)
	if err != nil {
		logError("failed to read config ", err)
		_, file, line, _ := runtime.Caller(0)
		fmt.Printf("Current line number: %s:%d\n", file, line)
		t.FailNow()
	}

	err = json.Unmarshal(c, &config)
	if err != nil {
		logError("failed to : ", err)
		_, file, line, _ := runtime.Caller(0)
		fmt.Printf("Current line number: %s:%d\n", file, line)
		t.FailNow()
	}
	fmt.Println(config)
	cs := S3Credintials{
		AccessKey: config.Bucket_Access_Key,
		SecretKey: config.Bucket_Secret_Key,
		URL:       config.Bucket_Url,
	}
	objectConfirmConnection(cs)
	fmt.Println("pass arguments to function")
	fmt.Println("return result")
}
