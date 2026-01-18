package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
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
	// Parse the multipart form with a max memory of 10 MB
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		utility.DefaultLogger.Error("parse form", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve the file from the parsed multipart form using the key "file"
	file, header, err := r.FormFile("file")
	if err != nil {
		utility.DefaultLogger.Error("parse file", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	d := db.ConfigDB(c)
	_, err = d.GetMediaByName(header.Filename)
	if err == nil {
		e := fmt.Errorf("duplicate entry found for %s\n", header.Filename)
		utility.DefaultLogger.Error("", e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
	forms := db.CreateMediaFormParams{
		Name:     header.Filename,
		AuthorID: "1",
	}
	params := db.MapCreateMediaParams(forms)

	row := d.CreateMedia(params)

	tmp, err := os.MkdirTemp(".", "temp")
	if err != nil {
		utility.DefaultLogger.Error("create tmp dir", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer exec.Command("rm", "-r", tmp)

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
