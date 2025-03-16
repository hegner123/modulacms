package bucket

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"testing"

	config "github.com/hegner123/modulacms/internal/Config"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

func TestObjectStorage(t *testing.T) {
	config := config.Config{}

	file, err := os.Open("testing-config.json")
	if err != nil {
		utility.LogError("failed to open config ", err)
		_, file, line, _ := runtime.Caller(0)
		fmt.Printf("Current line number: %s:%d\n", file, line)
		t.FailNow()
	}
	defer file.Close()
	c, err := io.ReadAll(file)
	if err != nil {
		utility.LogError("failed to read config ", err)
		_, file, line, _ := runtime.Caller(0)
		fmt.Printf("Current line number: %s:%d\n", file, line)
		t.FailNow()
	}

	err = json.Unmarshal(c, &config)
	if err != nil {
		utility.LogError("failed to : ", err)
		_, file, line, _ := runtime.Caller(0)
		fmt.Printf("Current line number: %s:%d\n", file, line)
		t.FailNow()
	}
	S3Access := S3Credintials{
		AccessKey: config.Bucket_Access_Key,
		SecretKey: config.Bucket_Secret_Key,
		URL:       config.Bucket_Url,
	}
	bucket, err := S3Access.GetBucket()
    if err!=nil {
        return
    }
	if bucket == nil {
		t.FailNow()

	}
}

func TestUpload(t *testing.T) {
	config := config.Config{}


	file, err := os.Open("testing-config.json")
	if err != nil {
		utility.LogError("failed to open config ", err)
		_, file, line, _ := runtime.Caller(0)
		fmt.Printf("Current line number: %s:%d\n", file, line)
		t.FailNow()
	}
	defer file.Close()
	c, err := io.ReadAll(file)
	if err != nil {
		utility.LogError("failed to read config ", err)
		_, file, line, _ := runtime.Caller(0)
		fmt.Printf("Current line number: %s:%d\n", file, line)
		t.FailNow()
	}
	file, err = os.Open("testFiles/test1.png")
	if err != nil {
		utility.LogError("failed to open File", err)
		_, file, line, _ := runtime.Caller(0)
		fmt.Printf("Current line number: %s:%d\n", file, line)
		t.FailNow()
	}
	err = json.Unmarshal(c, &config)
	if err != nil {
		utility.LogError("failed to : ", err)
		_, file, line, _ := runtime.Caller(0)
		fmt.Printf("Current line number: %s:%d\n", file, line)
		t.FailNow()
	}
	S3Access := S3Credintials{
		AccessKey: config.Bucket_Access_Key,
		SecretKey: config.Bucket_Secret_Key,
		URL:       config.Bucket_Url,
	}
	bucket, err := S3Access.GetBucket()
    if err!=nil {
        return
    }
	payload, err := UploadPrep("media/test1.png", "backups", file)
	if err != nil {
		utility.LogError("failed to : ", err)
		_, file, line, _ := runtime.Caller(0)
		fmt.Printf("Current line number: %s:%d\n", file, line)
		t.FailNow()
	}
	_, err = ObjectUpload(bucket, payload)
	if err != nil {
		utility.LogError("failed to : ", err)
	}
}
