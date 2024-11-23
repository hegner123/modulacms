package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

func adminRouter(w http.ResponseWriter, r *http.Request) {
	fmt.Print(r.URL.Path)
	switch r.URL.Path {
	case "/admin/field/add":
		fmt.Print("/admin/field/add\n")
		adminCreateField(w, r)
	case "/admin/media/create":
		fmt.Print("/admin/media/create\n")
		adminHandleUploadWithProgress(w, r)
	}
}

func adminCreateField(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("admin create field\n")
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	var field = mdb.Field{}

	field.ID = 0
	field.Routeid = 0
	form := r.ParseForm()
	fmt.Print(form)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"message": "wip"})
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func adminHandleUploadWithProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Transfer-Encoding", "chunked")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	err := r.ParseMultipartForm(sizeInBytes(1, GB))
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	uploadDir := "./tmp"
	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		logError("failed to create upload directory ", err)
	}
	tmpPath := filepath.Join(uploadDir, handler.Filename)
	dst, err := os.Create(tmpPath)
	if err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	totalBytes := r.ContentLength
	uploadedBytes := int64(0)

	buf := make([]byte, sizeInBytes(1, KB))
	for {
		n, err := file.Read(buf)
		if n > 0 {
			if _, writeErr := dst.Write(buf[:n]); writeErr != nil {
				http.Error(w, "Error saving the file", http.StatusInternalServerError)
				return
			}

			uploadedBytes += int64(n)
			progress := (float64(uploadedBytes) / float64(totalBytes)) * 100

			fmt.Fprintf(w, "Progress: %.2f%%\n", progress)
			flusher.Flush()
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			http.Error(w, "Error reading the file", http.StatusInternalServerError)
			return
		}
	}

	fmt.Fprint(w, "Upload complete!\n")
	flusher.Flush()
	handleCompletedMediaUpload(tmpPath, tmpPath)
}
