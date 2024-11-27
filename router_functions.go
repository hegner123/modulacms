package main

import (
	"encoding/json"
	"net/http"
)

func checkPath(segments []string, index Segment, ref string) bool {
    return segments[index]==ref
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

func formMapJson(r *http.Request) []byte {
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
