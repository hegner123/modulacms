package media

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	bucket "github.com/hegner123/modulacms/internal/bucket"
	config "github.com/hegner123/modulacms/internal/config"
	db "github.com/hegner123/modulacms/internal/db"
	utility "github.com/hegner123/modulacms/internal/utility"
)

//srcFile is the source file
//dstPath is the local destination for optimized files before uploading
func HandleMediaUpload(srcFile string, dstPath string, c config.Config) error {
	bucketDir := c.Bucket_Media
	now := time.Now()
	year := now.Year()
	month := now.Month()
    trimmedPrefix := strings.Split(srcFile, "/")
    last := len(trimmedPrefix)
	baseName := trimmedPrefix[last-1]

	optimized, err := OptimizeUpload(srcFile, dstPath, c)
	if err != nil {
		return err
	}
	s3Creds := bucket.S3Credintials{
		AccessKey: c.Bucket_Access_Key,
		SecretKey: c.Bucket_Secret_Key,
		URL:       c.Bucket_Endpoint,
	}
    
	s3 := s3Creds
	s3Session, err := s3.GetBucket()
	if err != nil {
		return err
	}
	srcset := []string{}
	for _, f := range *optimized {
		file, err := os.Open(f)
		if err != nil {
			return err
		}
		newPath := fmt.Sprintf("%s/%d/%d/%s", bucketDir, year, month, f)
		uploadPath := fmt.Sprintf("https://%s%s", c.Bucket_Endpoint, newPath)

		prep, err := bucket.UploadPrep(newPath, "media", file)
		if err != nil {
			return err
		}

		_, err = bucket.ObjectUpload(s3Session, prep)
		if err != nil {
			return err
		}
		srcset = append(srcset, uploadPath)

	}
	newSrcSet, err := json.Marshal(srcset)
	if err != nil {
		return err
	}
	d := db.ConfigDB(c)
    utility.DefaultLogger.Debug("SQL Filter Condition:", srcFile)
	rowPtr, err := d.GetMediaByName(baseName)
	if err != nil {
		return err
	}
	row := *rowPtr

	params := MapMediaParams(row)
	params.Srcset = db.StringToNullString(string(newSrcSet))
	_, err = d.UpdateMedia(params)
	if err != nil {
		return err
	}

	return nil
}

func MapMediaParams(a db.Media) db.UpdateMediaParams {
	return db.UpdateMediaParams{
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		Url:          a.Url,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		Srcset:       a.Srcset,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: db.StringToNullString(utility.TimestampReadable()),
		MediaID:      a.MediaID,
	}

}
