package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// ContentCreateRequest is the JSON body for POST /api/v1/content/create.
type ContentCreateRequest struct {
	ParentID   types.NullableContentID  `json:"parent_id"`
	RouteID    types.NullableRouteID    `json:"route_id"`
	DatatypeID types.DatatypeID         `json:"datatype_id"`
	Status     types.ContentStatus      `json:"status"`
	Fields     map[types.FieldID]string `json:"fields"`
}

// ContentCreateResponse is the JSON response for POST /api/v1/content/create.
type ContentCreateResponse struct {
	ContentData   *db.ContentData    `json:"content_data"`
	Fields        []db.ContentFields `json:"fields"`
	FieldsCreated int                `json:"fields_created"`
	FieldsFailed  int                `json:"fields_failed"`
	Errors        []string           `json:"errors"`
}

// ContentCreateHandler handles POST /api/v1/content/create.
func ContentCreateHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	apiContentCreate(w, r, svc)
}

// apiContentCreate creates a content node and all its datatype fields atomically.
func apiContentCreate(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req ContentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if err := req.DatatypeID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid datatype_id: %v", err), http.StatusBadRequest)
		return
	}

	d := svc.Driver()

	// Validate datatype exists.
	_, err := d.GetDatatype(req.DatatypeID)
	if err != nil {
		http.Error(w, fmt.Sprintf("datatype not found: %v", err), http.StatusBadRequest)
		return
	}

	// Derive author from authenticated user.
	user := middleware.AuthenticatedUser(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	now := types.TimestampNow()
	status := req.Status
	if status == "" {
		status = types.ContentStatusDraft
	}

	ctx := r.Context()
	cfg, err := svc.Config()
	if err != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *cfg)

	// Create the content node.
	createParams := db.CreateContentDataParams{
		ParentID:     req.ParentID,
		RouteID:      req.RouteID,
		DatatypeID:   types.NullableDatatypeID{ID: req.DatatypeID, Valid: true},
		AuthorID:     user.UserID,
		Status:       status,
		DateCreated:  now,
		DateModified: now,
	}
	contentData, err := d.CreateContentData(ctx, ac, createParams)
	if err != nil {
		utility.DefaultLogger.Error("content/create: failed to create content data", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Link into sibling chain if it has a parent.
	if contentData.ParentID.Valid {
		updated, chainErr := appendToSiblingChain(ctx, ac, d, contentData,
			types.NullableContentID{}, types.NullableContentID{})
		if chainErr != nil {
			utility.DefaultLogger.Error("content/create: sibling chain error", chainErr)
			http.Error(w, chainErr.Error(), http.StatusInternalServerError)
			return
		}
		contentData = updated
	}

	// Fetch fields for this datatype.
	fieldList, err := d.ListFieldsByDatatypeID(types.NullableDatatypeID{ID: req.DatatypeID, Valid: true})
	if err != nil {
		utility.DefaultLogger.Error("content/create: failed to list datatype fields", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := ContentCreateResponse{
		ContentData: contentData,
		Fields:      make([]db.ContentFields, 0),
		Errors:      make([]string, 0),
	}

	if fieldList != nil {
		for _, field := range *fieldList {
			// Use provided value or empty string.
			value := ""
			if req.Fields != nil {
				if v, ok := req.Fields[field.FieldID]; ok {
					value = v
				}
			}

			created, cfErr := d.CreateContentField(ctx, ac, db.CreateContentFieldParams{
				RouteID:       contentData.RouteID,
				ContentDataID: types.NullableContentID{ID: contentData.ContentDataID, Valid: true},
				FieldID:       types.NullableFieldID{ID: field.FieldID, Valid: true},
				FieldValue:    value,
				AuthorID:      user.UserID,
				DateCreated:   now,
				DateModified:  now,
			})
			if cfErr != nil {
				resp.FieldsFailed++
				resp.Errors = append(resp.Errors, fmt.Sprintf("field %s: %v", field.FieldID, cfErr))
				utility.DefaultLogger.Error(fmt.Sprintf("content/create: failed to create content field for field %s", field.FieldID), cfErr)
				continue
			}
			resp.FieldsCreated++
			resp.Fields = append(resp.Fields, *created)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
