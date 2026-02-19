package modula

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestResource creates a Resource[Datatype, ...] backed by a httptest server.
// The caller provides the server handler and receives the resource for testing.
func newTestResource(t *testing.T, handler http.HandlerFunc) (*Resource[Datatype, CreateDatatypeParams, UpdateDatatypeParams, DatatypeID], *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	h := &httpClient{
		baseURL:    srv.URL,
		httpClient: srv.Client(),
	}
	r := &Resource[Datatype, CreateDatatypeParams, UpdateDatatypeParams, DatatypeID]{
		path: "/api/v1/datatype",
		http: h,
	}
	return r, srv
}

func TestResource_List(t *testing.T) {
	datatypes := []Datatype{
		{DatatypeID: "dt-001", Label: "Article", Type: "content"},
		{DatatypeID: "dt-002", Label: "Page", Type: "content"},
	}

	res, srv := newTestResource(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/api/v1/datatype" {
			t.Errorf("path = %q, want /api/v1/datatype", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(datatypes)
	})
	defer srv.Close()

	result, err := res.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
	if result[0].DatatypeID != "dt-001" {
		t.Errorf("result[0].DatatypeID = %q, want %q", result[0].DatatypeID, "dt-001")
	}
	if result[1].Label != "Page" {
		t.Errorf("result[1].Label = %q, want %q", result[1].Label, "Page")
	}
}

func TestResource_Get(t *testing.T) {
	dt := Datatype{DatatypeID: "dt-001", Label: "Article", Type: "content"}

	res, srv := newTestResource(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/api/v1/datatype/" {
			t.Errorf("path = %q, want /api/v1/datatype/", r.URL.Path)
		}
		q := r.URL.Query().Get("q")
		if q != "dt-001" {
			t.Errorf("query param q = %q, want %q", q, "dt-001")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dt)
	})
	defer srv.Close()

	result, err := res.Get(context.Background(), DatatypeID("dt-001"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DatatypeID != "dt-001" {
		t.Errorf("DatatypeID = %q, want %q", result.DatatypeID, "dt-001")
	}
	if result.Label != "Article" {
		t.Errorf("Label = %q, want %q", result.Label, "Article")
	}
}

func TestResource_Create(t *testing.T) {
	res, srv := newTestResource(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/api/v1/datatype" {
			t.Errorf("path = %q, want /api/v1/datatype", r.URL.Path)
		}

		var params CreateDatatypeParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if params.Label != "BlogPost" {
			t.Errorf("params.Label = %q, want %q", params.Label, "BlogPost")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Datatype{
			DatatypeID: "dt-new",
			Label:      params.Label,
			Type:       params.Type,
		})
	})
	defer srv.Close()

	result, err := res.Create(context.Background(), CreateDatatypeParams{
		Label: "BlogPost",
		Type:  "content",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DatatypeID != "dt-new" {
		t.Errorf("DatatypeID = %q, want %q", result.DatatypeID, "dt-new")
	}
	if result.Label != "BlogPost" {
		t.Errorf("Label = %q, want %q", result.Label, "BlogPost")
	}
}

func TestResource_Update(t *testing.T) {
	res, srv := newTestResource(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %q, want PUT", r.Method)
		}
		if r.URL.Path != "/api/v1/datatype/" {
			t.Errorf("path = %q, want /api/v1/datatype/", r.URL.Path)
		}

		var params UpdateDatatypeParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Datatype{
			DatatypeID: params.DatatypeID,
			Label:      params.Label,
			Type:       params.Type,
		})
	})
	defer srv.Close()

	result, err := res.Update(context.Background(), UpdateDatatypeParams{
		DatatypeID: "dt-001",
		Label:      "Updated Article",
		Type:       "content",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Label != "Updated Article" {
		t.Errorf("Label = %q, want %q", result.Label, "Updated Article")
	}
}

func TestResource_Delete(t *testing.T) {
	res, srv := newTestResource(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		if r.URL.Path != "/api/v1/datatype/" {
			t.Errorf("path = %q, want /api/v1/datatype/", r.URL.Path)
		}
		q := r.URL.Query().Get("q")
		if q != "dt-del" {
			t.Errorf("query param q = %q, want %q", q, "dt-del")
		}
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	err := res.Delete(context.Background(), DatatypeID("dt-del"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResource_ListPaginated(t *testing.T) {
	res, srv := newTestResource(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}

		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")
		if limit != "10" {
			t.Errorf("limit = %q, want %q", limit, "10")
		}
		if offset != "20" {
			t.Errorf("offset = %q, want %q", offset, "20")
		}

		resp := PaginatedResponse[Datatype]{
			Data: []Datatype{
				{DatatypeID: "dt-021", Label: "Item21", Type: "content"},
				{DatatypeID: "dt-022", Label: "Item22", Type: "content"},
			},
			Total:  50,
			Limit:  10,
			Offset: 20,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	result, err := res.ListPaginated(context.Background(), PaginationParams{Limit: 10, Offset: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 50 {
		t.Errorf("Total = %d, want %d", result.Total, 50)
	}
	if result.Limit != 10 {
		t.Errorf("Limit = %d, want %d", result.Limit, 10)
	}
	if result.Offset != 20 {
		t.Errorf("Offset = %d, want %d", result.Offset, 20)
	}
	if len(result.Data) != 2 {
		t.Fatalf("len(Data) = %d, want 2", len(result.Data))
	}
	if result.Data[0].DatatypeID != "dt-021" {
		t.Errorf("Data[0].DatatypeID = %q, want %q", result.Data[0].DatatypeID, "dt-021")
	}
}

func TestResource_Count(t *testing.T) {
	res, srv := newTestResource(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		countParam := r.URL.Query().Get("count")
		if countParam != "true" {
			t.Errorf("count = %q, want %q", countParam, "true")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int64{"count": 42})
	})
	defer srv.Close()

	count, err := res.Count(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 42 {
		t.Errorf("count = %d, want %d", count, 42)
	}
}

func TestResource_NotFound(t *testing.T) {
	res, srv := newTestResource(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, `{"message":"datatype not found"}`)
	})
	defer srv.Close()

	_, err := res.Get(context.Background(), DatatypeID("nonexistent"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound() = false, want true; error = %v", err)
	}

	apiErr, ok := err.(*ApiError)
	if ok {
		t.Errorf("direct type assertion should fail because error is wrapped; got *ApiError directly")
		_ = apiErr
	}
	// The error is wrapped by fmt.Errorf in Resource.Get, so use errors.As via IsNotFound
}
