package main

import (
	"net/http"
	"net/url"
	"strings"
)

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
