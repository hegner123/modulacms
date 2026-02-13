# bucket

S3-compatible object storage integration for ModulaCMS.

## Overview

The bucket package provides S3-compatible object storage operations using the AWS SDK. It handles credential management, session creation, and file upload operations for any S3-compatible service including AWS S3, MinIO, and other compatible storage providers. The package supports configurable endpoints, regions, and path-style access for maximum compatibility.

### Types

#### type Metadata

```go
type Metadata map[string]string
```

Metadata represents key-value pairs attached to S3 objects. Use this type to store custom metadata that should be associated with uploaded files.

#### type S3Credentials

```go
type S3Credentials struct {
    AccessKey      string
    SecretKey      string
    URL            string
    Region         string
    ForcePathStyle bool
}
```

S3Credentials holds authentication and configuration for S3-compatible storage. AccessKey and SecretKey provide authentication. URL specifies the storage endpoint. Region defaults to us-east-1 if empty. ForcePathStyle enables path-style bucket access required by some S3-compatible services like MinIO.

### Functions

#### func GetS3Creds

```go
func GetS3Creds(c *config.Config) *S3Credentials
```

GetS3Creds extracts S3 credentials from application configuration and returns an S3Credentials struct. This function reads bucket configuration fields from the config object and constructs the credentials needed to establish an S3 session.

```go
creds := bucket.GetS3Creds(appConfig)
svc, err := creds.GetBucket()
```

#### func (S3Credentials) GetBucket

```go
func (cs S3Credentials) GetBucket() (*s3.S3, error)
```

GetBucket creates an AWS S3 session using the credentials and returns an S3 service client. If region is empty, it defaults to us-east-1. Returns an error if session creation fails due to invalid credentials or unreachable endpoint.

```go
creds := bucket.S3Credentials{
    AccessKey: "access",
    SecretKey: "secret",
    URL: "https://s3.amazonaws.com",
}
svc, err := creds.GetBucket()
if err != nil {
    log.Fatal(err)
}
```

#### func UploadPrep

```go
func UploadPrep(uploadPath string, bucketName string, data *os.File, acl string) (*s3.PutObjectInput, error)
```

UploadPrep constructs a PutObjectInput for uploading a file to S3. uploadPath is the destination key in the bucket. bucketName is the target bucket name. data is an open file handle to upload. acl sets object permissions like public-read or private. Returns the prepared input struct ready for ObjectUpload.

```go
file, _ := os.Open("/path/to/file.jpg")
defer file.Close()
input, err := bucket.UploadPrep("uploads/file.jpg", "my-bucket", file, "public-read")
```

#### func ObjectUpload

```go
func ObjectUpload(s3 *s3.S3, payload *s3.PutObjectInput) (*s3.PutObjectOutput, error)
```

ObjectUpload executes the file upload to S3 using the service client and prepared payload. The s3 argument must be obtained from GetBucket. The payload must be created by UploadPrep. Returns upload metadata on success. Returns an error if the upload fails due to network issues, invalid credentials, or bucket permissions.

```go
svc, _ := creds.GetBucket()
input, _ := bucket.UploadPrep("file.jpg", "bucket", file, "private")
output, err := bucket.ObjectUpload(svc, input)
if err != nil {
    log.Printf("Upload failed: %v", err)
}
```

#### func PrintBuckets

```go
func PrintBuckets(s3 *s3.S3)
```

PrintBuckets lists all buckets accessible to the S3 client and prints their names and creation dates to the logger. This function is primarily for debugging and administrative tasks. It terminates the program if bucket listing fails.

```go
svc, _ := creds.GetBucket()
bucket.PrintBuckets(svc)
```
