package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	mdb "github.com/hegner123/modulacms/db-sqlite"
)

type Metadata map[string]string

func (cs S3Credintials) getBucket() *s3.S3 {
	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(cs.AccessKey, cs.SecretKey, ""),
		Endpoint:         aws.String(cs.URL),
		Region:           aws.String("us-southeast-1"), // Use any valid AWS region
		S3ForcePathStyle: aws.Bool(true),               // Required for Linode Object Storage
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	return s3.New(sess)
}

func UploadPrep(uploadPath string, bucketName string, data *os.File) (*s3.PutObjectInput, error) {
	upload := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(uploadPath),
		Body:   data,
		ACL:    aws.String("public-read"),
	}
	return upload, nil
}

func PrintBuckets(s3 *s3.S3) {
	// Example: List buckets
	result, err := s3.ListBuckets(nil)
	if err != nil {
		log.Fatalf("Unable to list buckets: %v", err)
	}

	fmt.Println("Buckets:")
	for _, bucket := range result.Buckets {
		fmt.Printf("%s created on %s\n",
			aws.StringValue(bucket.Name),
			aws.TimeValue(bucket.CreationDate))
	}
}

func ObjectUpload(s3 *s3.S3, payload *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	upload, err := s3.PutObject(payload)
	if err != nil {
		logError("failed to upload ", err)
	}
	return upload, nil
}

func ParseMetaData(dbEntry mdb.Media) {
}
