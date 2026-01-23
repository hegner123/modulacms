package router

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/media"
	"github.com/hegner123/modulacms/internal/utility"
)

func MediaUploadHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	fmt.Println("upload Route")
	switch r.Method {
	case http.MethodPost:
		apiCreateMediaUpload(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func apiCreateMediaUpload(w http.ResponseWriter, r *http.Request, c config.Config) {
	const MaxFileSize = 10 << 20 // 10 MB

	// Parse the multipart form with a max memory of 10 MB
	err := r.ParseMultipartForm(MaxFileSize)
	if err != nil {
		utility.DefaultLogger.Error("parse form", err)
		http.Error(w, "File too large or invalid multipart form", http.StatusBadRequest)
		return
	}

	// Retrieve the file from the parsed multipart form using the key "file"
	file, header, err := r.FormFile("file")
	if err != nil {
		utility.DefaultLogger.Error("parse file", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Explicit file size validation
	if header.Size > MaxFileSize {
		utility.DefaultLogger.Error("file too large", fmt.Errorf("size: %d", header.Size))
		http.Error(w, fmt.Sprintf("File size %d exceeds maximum %d", header.Size, MaxFileSize), http.StatusBadRequest)
		return
	}

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		utility.DefaultLogger.Error("read file header", err)
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Reset file pointer
	_, err = file.Seek(0, 0)
	if err != nil {
		utility.DefaultLogger.Error("seek file", err)
		http.Error(w, "Failed to process file", http.StatusInternalServerError)
		return
	}

	// Validate MIME type
	contentType := http.DetectContentType(buffer)
	validTypes := map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
		"image/gif":  true,
		"image/webp": true,
	}

	if !validTypes[contentType] {
		utility.DefaultLogger.Error("invalid content type", fmt.Errorf("type: %s", contentType))
		http.Error(w, fmt.Sprintf("Invalid file type: %s. Only images allowed.", contentType), http.StatusBadRequest)
		return
	}

	// Note: WebP images can be decoded but cannot be encoded.
	// WebP uploads will fail during optimization if WebP encoding is attempted.
	// To support WebP output, add a WebP encoding library (e.g., github.com/chai2010/webp)
	if contentType == "image/webp" {
		utility.DefaultLogger.Info("WebP upload detected - WebP encoding not supported, may fail during optimization")
	}

	d := db.ConfigDB(c)
	_, err = d.GetMediaByName(header.Filename)
	if err == nil {
		e := fmt.Errorf("duplicate entry found for %s\n", header.Filename)
		utility.DefaultLogger.Error("", e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
	params := db.CreateMediaParams{
		Name:         sql.NullString{String: header.Filename, Valid: true},
		AuthorID:     types.NullableUserID{ID: types.UserID("1"), Valid: true}, // TODO: Get from authenticated session
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	}

	row := d.CreateMedia(params)

	tmp, err := os.MkdirTemp("", "modulacms-media")
	if err != nil {
		utility.DefaultLogger.Error("create tmp dir", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmp)

	// Create a destination file on the server
	dst, err := os.Create(tmp + "/" + header.Filename)
	if err != nil {
		utility.DefaultLogger.Error("create destination", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file data to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		utility.DefaultLogger.Error("copy file", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = media.HandleMediaUpload(tmp+"/"+header.Filename, tmp, c)
	if err != nil {
		utility.DefaultLogger.Error("handle media upload", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(row)

}
