package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/model"
	"github.com/hegner123/modulacms/internal/utility"
)

// GlobalEntry represents a single global content tree in the delivery response.
type GlobalEntry struct {
	ContentDataID string     `json:"content_data_id"`
	DatatypeID    string     `json:"datatype_id"`
	DatatypeName  string     `json:"datatype_name"`
	DatatypeLabel string     `json:"datatype_label"`
	Tree          model.Node `json:"tree"`
}

// GlobalsHandler serves all _global content trees for public delivery.
func GlobalsHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := apiGetGlobals(w, r, c); err != nil {
		utility.DefaultLogger.Error("globals handler error", err)
	}
}

// apiGetGlobals fetches all _global root content data, builds a tree for each,
// and returns them as an array.
func apiGetGlobals(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	globals, err := d.ListContentDataGlobal()
	if err != nil {
		http.Error(w, "failed to list globals", http.StatusInternalServerError)
		return fmt.Errorf("ListContentDataGlobal: %w", err)
	}

	if globals == nil || len(*globals) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]GlobalEntry{})
		return nil
	}

	entries := make([]GlobalEntry, 0, len(*globals))

	for _, root := range *globals {
		entry, err := buildGlobalEntry(r.Context(), d, root)
		if err != nil {
			utility.DefaultLogger.Warn("skipping global entry", err)
			continue
		}
		entries = append(entries, *entry)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(entries)
}

// buildGlobalEntry builds a tree for a single _global root content data node.
func buildGlobalEntry(ctx context.Context, d db.DbDriver, root db.ContentData) (*GlobalEntry, error) {
	// Collect all content data in the subtree under this root (recursive CTE).
	allCD, err := d.GetContentDataDescendants(ctx, root.ContentDataID)
	if err != nil {
		return nil, fmt.Errorf("subtree for %s: %w", root.ContentDataID, err)
	}

	// Build parallel slices of content data + datatypes.
	cdSlice := *allCD
	dtSlice := make([]db.Datatypes, 0, len(cdSlice))
	for _, cd := range cdSlice {
		if !cd.DatatypeID.Valid {
			dtSlice = append(dtSlice, db.Datatypes{})
			continue
		}
		dt, err := d.GetDatatype(cd.DatatypeID.ID)
		if err != nil {
			return nil, fmt.Errorf("get datatype %s: %w", cd.DatatypeID.ID, err)
		}
		dtSlice = append(dtSlice, *dt)
	}

	// Collect all content fields for nodes in this subtree.
	var allCF []db.ContentFields
	var allFD []db.Fields
	for _, cd := range cdSlice {
		ncid := types.NullableContentID{ID: cd.ContentDataID, Valid: true}
		fields, err := d.ListContentFieldsByContentData(ncid)
		if err != nil {
			return nil, fmt.Errorf("fields for %s: %w", cd.ContentDataID, err)
		}
		if fields == nil {
			continue
		}
		for _, cf := range *fields {
			if !cf.FieldID.Valid {
				continue
			}
			fd, err := d.GetField(cf.FieldID.ID)
			if err != nil {
				return nil, fmt.Errorf("get field %s: %w", cf.FieldID.ID, err)
			}
			allCF = append(allCF, cf)
			allFD = append(allFD, *fd)
		}
	}

	built, err := model.BuildTree(utility.DefaultLogger, cdSlice, dtSlice, allCF, allFD)
	if err != nil {
		return nil, fmt.Errorf("BuildTree for global %s: %w", root.ContentDataID, err)
	}

	if built.Node == nil {
		return nil, fmt.Errorf("no root node built for global %s", root.ContentDataID)
	}

	// Resolve datatype metadata for the response.
	var dtName, dtLabel string
	if root.DatatypeID.Valid {
		dt, err := d.GetDatatype(root.DatatypeID.ID)
		if err == nil {
			dtName = dt.Name
			dtLabel = dt.Label
		}
	}

	return &GlobalEntry{
		ContentDataID: root.ContentDataID.String(),
		DatatypeID:    root.DatatypeID.ID.String(),
		DatatypeName:  dtName,
		DatatypeLabel: dtLabel,
		Tree:          *built.Node,
	}, nil
}
