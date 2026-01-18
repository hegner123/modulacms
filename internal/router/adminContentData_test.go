package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	config "github.com/hegner123/modulacms/internal/config"
	db "github.com/hegner123/modulacms/internal/db"
)

// TestApiCreateAdminContentData_Success tests a successful POST request.
func TestApiCreateAdminContentData_Success(t *testing.T) {
	p := config.NewFileProvider("")
	m := config.NewManager(p)
	c, err := m.Config()
	if err != nil {
		t.Fatal(err)
	}
	// Initialize a dummy config.

	// Create dummy admin content data.
	// Replace "Title" and "Body" with the actual fields of CreateAdminContentDataParams.
	newContent := db.CreateAdminContentDataParams{
		AdminRouteID:    1,
		AdminDatatypeID: 1,
		History:         db.StringToNullString(""),
		DateCreated:     db.StringToNullString(fmt.Sprint(time.Now().Unix())),
		DateModified:    db.StringToNullString(fmt.Sprint(time.Now().Unix())),
	}

	// Encode the dummy data to JSON.
	requestBody, err := json.Marshal(newContent)
	if err != nil {
		t.Fatalf("Failed to marshal request data: %v", err)
	}

	// Create a new HTTP request with the JSON body.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admincontent", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder to capture the handler's response.
	rr := httptest.NewRecorder()

	// Call the handler function.
	err = apiCreateAdminContentData(rr, req, *c)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify that the status code is 201 Created.
	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, rr.Code)
	}

	// Verify that the Content-Type header is set to "application/json".
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", ct)
	}

	// Optionally, decode the response body to check the returned data.
	var createdData db.CreateAdminContentDataParams
	if err := json.NewDecoder(rr.Body).Decode(&createdData); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

}

/*
// TestApiCreateAdminContentData_InvalidJSON tests the handler with an invalid JSON body.
func TestApiCreateAdminContentData_InvalidJSON(t *testing.T) {
	cfg := config.Config{
		// Minimal config initialization.
	}

	// Create a request with invalid JSON.
	req := httptest.NewRequest(http.MethodPost, "/admin-content", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Call the handler function.
	err := apiCreateAdminContentData(rr, req, cfg)
	if err == nil {
		t.Fatal("Expected error when decoding invalid JSON, got nil")
	}

	// Check that the status code is set to 400 Bad Request.
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d for invalid JSON, got %d", http.StatusBadRequest, rr.Code)
	}
}
*/
