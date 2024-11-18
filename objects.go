package main

import (
    "fmt"
    "log"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
)

func confirmConnection() {
    // Replace these with your actual credentials and endpoint
    accessKey := "YOUR_ACCESS_KEY"
    secretKey := "YOUR_SECRET_KEY"
    endpoint := "https://us-east-1.linodeobjects.com" // Example endpoint

    // Create a new session with the provided credentials and endpoint
    sess, err := session.NewSession(&aws.Config{
        Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
        Endpoint:         aws.String(endpoint),
        Region:           aws.String("us-east-1"), // Use any valid AWS region
        S3ForcePathStyle: aws.Bool(true),          // Required for Linode Object Storage
    })
    if err != nil {
        log.Fatalf("Failed to create session: %v", err)
    }

    // Create a new S3 service client
    svc := s3.New(sess)

    // Example: List buckets
    result, err := svc.ListBuckets(nil)
    if err != nil {
        log.Fatalf("Unable to list buckets: %v", err)
    }

    fmt.Println("Buckets:")
    for _, bucket := range result.Buckets {
        fmt.Printf("* %s created on %s\n",
            aws.StringValue(bucket.Name),
            aws.TimeValue(bucket.CreationDate))
    }
}
