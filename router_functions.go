package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

func checkPath(rawURL, match string) bool  {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	path := strings.Trim(parsedURL.Path, "/")

	segments := strings.Split(path, "/")

	if len(segments) > 0 && segments[0] == match {
		return true
	}

	return false
}

func checkAPIPath(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return strings.HasPrefix(parsedURL.Path, "/api/")
}

func matchesPath(text, searchTerm string) bool {
	return strings.Contains(text, searchTerm)
}
func parseQueryParams(r *http.Request) (map[string]string, error) {
	params := make(map[string]string)
    
	query := r.URL.Query()

	for key, values := range query {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	return params, nil
}

func stripAPIPath(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsedURL.Path = strings.TrimPrefix(parsedURL.Path, "/api/")
	return parsedURL.String(), nil
}
func stripDeletePath(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsedURL.Path = strings.TrimPrefix(parsedURL.Path, "/delete/")
	return parsedURL.String(), nil
}
func stripGetPath(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsedURL.Path = strings.TrimPrefix(parsedURL.Path, "/get/")
	return parsedURL.String(), nil
}
func stripListPath(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsedURL.Path = strings.TrimPrefix(parsedURL.Path, "/list/")
	return parsedURL.String(), nil
}
func stripCreatePath(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsedURL.Path = strings.TrimPrefix(parsedURL.Path, "/create/")
	return parsedURL.String(), nil
}
func stripUpdatePath(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsedURL.Path = strings.TrimPrefix(parsedURL.Path, "/update/")
	return parsedURL.String(), nil
}


func formMapJson(r *http.Request)[]byte{
	formData := make(map[string]string)
	for key, values := range r.Form {
		formData[key] = values[0]
	}

	jsonBytes, err := json.Marshal(formData)
    if err != nil { 
        logError("failed to Marshal JSON", err)
    }
    return jsonBytes
}
