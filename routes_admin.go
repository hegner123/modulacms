package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func adminRouter(w http.ResponseWriter, r *http.Request) {
	fmt.Print(r.URL.Path)
	switch r.URL.Path {
	case "/admin/field/add":
		fmt.Print("/admin/field/add\n")
		adminCreateField(w, r)
	case "/admin/media/create":
		fmt.Print("/admin/media/create\n")
		adminUploadMedia(w, r)
	}
}

func adminCreateField(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("admin create field\n")
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	var field = Field{}

	field.ID = 0
	field.RouteID = 0
	form := r.ParseForm()
	fmt.Print(form)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"message": "wip"})
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func adminUploadMedia(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(int64(1000))
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	file, header, err := r.FormFile("media")
	if err != nil {
		http.Error(w, "Unable to get the file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	var buffer bytes.Buffer

	_, err = io.Copy(&buffer, file)
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusInternalServerError)
		return
	}

	fmt.Println("File content as bytes.Buffer:", buffer.String())
    handleMediaUpload(&buffer,header.Filename)

}
