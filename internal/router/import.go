package router

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/model"
	"github.com/hegner123/modulacms/internal/transform"
	"github.com/hegner123/modulacms/internal/utility"
)

// ImportContentfulHandler handles importing Contentful format to ModulaCMS
func ImportContentfulHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiImportContent(w, r, c, config.FormatContentful)
}

// ImportSanityHandler handles importing Sanity format to ModulaCMS
func ImportSanityHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiImportContent(w, r, c, config.FormatSanity)
}

// ImportStrapiHandler handles importing Strapi format to ModulaCMS
func ImportStrapiHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiImportContent(w, r, c, config.FormatStrapi)
}

// ImportWordPressHandler handles importing WordPress format to ModulaCMS
func ImportWordPressHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiImportContent(w, r, c, config.FormatWordPress)
}

// ImportCleanHandler handles importing Clean ModulaCMS format
func ImportCleanHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiImportContent(w, r, c, config.FormatClean)
}

// ImportBulkHandler handles bulk import with format specified in request
func ImportBulkHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get format from query parameter
	format := r.URL.Query().Get("format")
	if format == "" {
		http.Error(w, "format query parameter required", http.StatusBadRequest)
		return
	}

	if !config.IsValidOutputFormat(format) {
		http.Error(w, "Invalid format. Valid options: contentful, sanity, strapi, wordpress, clean", http.StatusBadRequest)
		return
	}

	apiImportContent(w, r, c, config.OutputFormat(format))
}

// apiImportContent handles the core import logic
func apiImportContent(w http.ResponseWriter, r *http.Request, c config.Config, format config.OutputFormat) {
	d := db.ConfigDB(c)

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utility.DefaultLogger.Error("Failed to read request body", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Create transformer for the specified format
	transformCfg := transform.NewTransformConfig(
		format,
		c.Client_Site,
		c.Space_ID,
		d,
	)

	transformer, err := transformCfg.GetTransformer()
	if err != nil {
		utility.DefaultLogger.Error("Failed to get transformer", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse CMS format to ModulaCMS
	root, err := transformer.Parse(body)
	if err != nil {
		utility.DefaultLogger.Error("Failed to parse input", err)
		http.Error(w, "Failed to parse input: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Parse optional route_id from query parameter
	routeID := types.NullableRouteID{Valid: false}
	routeIDStr := r.URL.Query().Get("route_id")
	if routeIDStr != "" {
		routeID = types.NullableRouteID{
			ID:    types.RouteID(routeIDStr),
			Valid: true,
		}
	}

	// Import to database
	ac := middleware.AuditContextFromRequest(r, c)
	result, err := importRootToDatabase(r.Context(), ac, d, root, routeID)
	if err != nil {
		utility.DefaultLogger.Error("Failed to import to database", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	Success          bool     `json:"success"`
	DatatypesCreated int      `json:"datatypes_created"`
	FieldsCreated    int      `json:"fields_created"`
	ContentCreated   int      `json:"content_created"`
	Errors           []string `json:"errors,omitempty"`
	Message          string   `json:"message"`
}

// importContext tracks state during recursive import
type importContext struct {
	ctx           context.Context
	ac            audited.AuditContext
	driver        db.DbDriver
	routeID       types.NullableRouteID
	authorID      types.NullableUserID
	datatypeCache map[string]types.DatatypeID // "label|type" -> existing DatatypeID
	result        *ImportResult
}

// importRootToDatabase imports a parsed Root structure to the database
func importRootToDatabase(reqCtx context.Context, ac audited.AuditContext, driver db.DbDriver, root model.Root, routeID types.NullableRouteID) (*ImportResult, error) {
	result := &ImportResult{
		Success: false,
		Errors:  []string{},
	}

	if root.Node == nil {
		result.Errors = append(result.Errors, "No content to import")
		return result, nil
	}

	ctx := &importContext{
		ctx:           reqCtx,
		ac:            ac,
		driver:        driver,
		routeID:       routeID,
		authorID:      types.NullableUserID{Valid: false},
		datatypeCache: make(map[string]types.DatatypeID),
		result:        result,
	}

	ctx.importNode(root.Node, types.NullableContentID{Valid: false})

	result.Success = len(result.Errors) == 0
	if result.Success {
		result.Message = fmt.Sprintf("Import complete: %d datatypes, %d fields, %d content nodes created",
			result.DatatypesCreated, result.FieldsCreated, result.ContentCreated)
	} else {
		result.Message = fmt.Sprintf("Import completed with %d errors", len(result.Errors))
	}

	utility.DefaultLogger.Info("Import completed",
		"datatypes", result.DatatypesCreated,
		"fields", result.FieldsCreated,
		"content", result.ContentCreated,
		"errors", len(result.Errors),
	)

	return result, nil
}

// importNode recursively imports a single node and its children into the database.
// Returns the ContentID of the created content_data row (empty if creation failed).
func (ctx *importContext) importNode(node *model.Node, parentID types.NullableContentID) types.ContentID {
	// Find or create the datatype
	datatypeID := ctx.findOrCreateDatatype(node)
	if datatypeID.IsZero() {
		ctx.result.Errors = append(ctx.result.Errors,
			fmt.Sprintf("failed to create datatype for label=%q type=%q", node.Datatype.Info.Label, node.Datatype.Info.Type))
		return types.ContentID("")
	}

	// Create content_data with null sibling pointers (patched after children are created)
	now := types.TimestampNow()
	contentData, createErr := ctx.driver.CreateContentData(ctx.ctx, ctx.ac, db.CreateContentDataParams{
		RouteID:  ctx.routeID,
		ParentID: parentID,
		DatatypeID: types.NullableDatatypeID{
			ID:    datatypeID,
			Valid: true,
		},
		AuthorID:      ctx.authorID,
		Status:        types.ContentStatusDraft,
		FirstChildID:  sql.NullString{Valid: false},
		NextSiblingID: sql.NullString{Valid: false},
		PrevSiblingID: sql.NullString{Valid: false},
		DateCreated:   now,
		DateModified:  now,
	})
	if createErr != nil {
		ctx.result.Errors = append(ctx.result.Errors,
			fmt.Sprintf("failed to create content_data for node type=%q: %v", node.Datatype.Info.Label, createErr))
		return types.ContentID("")
	}

	contentDataID := contentData.ContentDataID
	if contentDataID.IsZero() {
		ctx.result.Errors = append(ctx.result.Errors,
			fmt.Sprintf("failed to create content_data for node type=%q", node.Datatype.Info.Label))
		return types.ContentID("")
	}

	// Create fields and content_fields for this node
	for _, field := range node.Fields {
		ctx.createFieldAndContentField(field, contentDataID, datatypeID)
	}

	// Recurse into children and collect their content IDs
	childIDs := make([]types.ContentID, 0, len(node.Nodes))
	for _, child := range node.Nodes {
		childContentID := ctx.importNode(child, types.NullableContentID{
			ID:    contentDataID,
			Valid: true,
		})
		if !childContentID.IsZero() {
			childIDs = append(childIDs, childContentID)
		}
	}

	// Patch sibling pointers after all children are created
	if len(childIDs) > 0 {
		ctx.patchSiblingPointers(*contentData, childIDs)
	}

	ctx.result.ContentCreated++
	return contentDataID
}

// findOrCreateDatatype returns an existing or newly-created DatatypeID for the node's type.
// De-duplicates by "label|type" key.
func (ctx *importContext) findOrCreateDatatype(node *model.Node) types.DatatypeID {
	label := node.Datatype.Info.Label
	typeStr := node.Datatype.Info.Type
	cacheKey := label + "|" + typeStr

	if existing, ok := ctx.datatypeCache[cacheKey]; ok {
		return existing
	}

	now := types.TimestampNow()
	created, createErr := ctx.driver.CreateDatatype(ctx.ctx, ctx.ac, db.CreateDatatypeParams{
		ParentID:     types.NullableContentID{Valid: false},
		Label:        label,
		Type:         typeStr,
		AuthorID:     ctx.authorID,
		DateCreated:  now,
		DateModified: now,
	})

	if createErr != nil || created.DatatypeID.IsZero() {
		return types.DatatypeID("")
	}

	ctx.datatypeCache[cacheKey] = created.DatatypeID
	ctx.result.DatatypesCreated++
	return created.DatatypeID
}

// createFieldAndContentField creates a field definition, a content_field linking
// it to the content_data, and a datatype_field linking the field to the datatype.
func (ctx *importContext) createFieldAndContentField(field model.Field, contentDataID types.ContentID, datatypeID types.DatatypeID) {
	now := types.TimestampNow()

	// Create the field definition
	createdField, fieldErr := ctx.driver.CreateField(ctx.ctx, ctx.ac, db.CreateFieldParams{
		ParentID: types.NullableDatatypeID{
			ID:    datatypeID,
			Valid: true,
		},
		Label:        field.Info.Label,
		Data:         "",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         types.FieldType(field.Info.Type),
		AuthorID:     ctx.authorID,
		DateCreated:  now,
		DateModified: now,
	})

	if fieldErr != nil || createdField.FieldID.IsZero() {
		ctx.result.Errors = append(ctx.result.Errors,
			fmt.Sprintf("failed to create field label=%q for content_data=%s", field.Info.Label, contentDataID))
		return
	}

	// Create the content_field linking content_data to field value
	_, cfErr := ctx.driver.CreateContentField(ctx.ctx, ctx.ac, db.CreateContentFieldParams{
		RouteID: ctx.routeID,
		ContentDataID: types.NullableContentID{
			ID:    contentDataID,
			Valid: true,
		},
		FieldID: types.NullableFieldID{
			ID:    createdField.FieldID,
			Valid: true,
		},
		FieldValue:   field.Content.FieldValue,
		AuthorID:     ctx.authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if cfErr != nil {
		ctx.result.Errors = append(ctx.result.Errors,
			fmt.Sprintf("failed to create content_field for field=%s content_data=%s: %v", createdField.FieldID, contentDataID, cfErr))
	}

	// Create the datatype_field linking the datatype to the field
	_, dtfErr := ctx.driver.CreateDatatypeField(ctx.ctx, ctx.ac, db.CreateDatatypeFieldParams{
		DatatypeID: types.NullableDatatypeID{
			ID:    datatypeID,
			Valid: true,
		},
		FieldID: types.NullableFieldID{
			ID:    createdField.FieldID,
			Valid: true,
		},
	})
	if dtfErr != nil {
		ctx.result.Errors = append(ctx.result.Errors,
			fmt.Sprintf("failed to create datatype_field for field=%s datatype=%s: %v", createdField.FieldID, datatypeID, dtfErr))
	}

	ctx.result.FieldsCreated++
}

// patchSiblingPointers sets first_child_id on the parent and links children
// with prev_sibling_id/next_sibling_id pointers.
//
// For children [c0, c1, c2]:
//
//	parent.FirstChildID = c0.ID
//	c0.PrevSiblingID = null,  c0.NextSiblingID = c1.ID
//	c1.PrevSiblingID = c0.ID, c1.NextSiblingID = c2.ID
//	c2.PrevSiblingID = c1.ID, c2.NextSiblingID = null
func (ctx *importContext) patchSiblingPointers(parent db.ContentData, childIDs []types.ContentID) {
	// Set first_child_id on the parent
	_, err := ctx.driver.UpdateContentData(ctx.ctx, ctx.ac, db.UpdateContentDataParams{
		ContentDataID: parent.ContentDataID,
		RouteID:       parent.RouteID,
		ParentID:      parent.ParentID,
		FirstChildID:  sql.NullString{String: childIDs[0].String(), Valid: true},
		NextSiblingID: parent.NextSiblingID,
		PrevSiblingID: parent.PrevSiblingID,
		DatatypeID:    parent.DatatypeID,
		AuthorID:      parent.AuthorID,
		Status:        parent.Status,
		DateCreated:   parent.DateCreated,
		DateModified:  parent.DateModified,
	})
	if err != nil {
		ctx.result.Errors = append(ctx.result.Errors,
			fmt.Sprintf("failed to set first_child_id on content_data=%s: %v", parent.ContentDataID, err))
	}

	// Link each child with sibling pointers
	for i, childID := range childIDs {
		prevSibling := sql.NullString{Valid: false}
		if i > 0 {
			prevSibling = sql.NullString{String: childIDs[i-1].String(), Valid: true}
		}

		nextSibling := sql.NullString{Valid: false}
		if i < len(childIDs)-1 {
			nextSibling = sql.NullString{String: childIDs[i+1].String(), Valid: true}
		}

		// We need the full row to update — fetch it first
		childData, fetchErr := ctx.driver.GetContentData(childID)
		if fetchErr != nil {
			ctx.result.Errors = append(ctx.result.Errors,
				fmt.Sprintf("failed to fetch content_data=%s for sibling patch: %v", childID, fetchErr))
			continue
		}

		_, updateErr := ctx.driver.UpdateContentData(ctx.ctx, ctx.ac, db.UpdateContentDataParams{
			ContentDataID: childData.ContentDataID,
			RouteID:       childData.RouteID,
			ParentID:      childData.ParentID,
			FirstChildID:  childData.FirstChildID,
			NextSiblingID: nextSibling,
			PrevSiblingID: prevSibling,
			DatatypeID:    childData.DatatypeID,
			AuthorID:      childData.AuthorID,
			Status:        childData.Status,
			DateCreated:   childData.DateCreated,
			DateModified:  childData.DateModified,
		})
		if updateErr != nil {
			ctx.result.Errors = append(ctx.result.Errors,
				fmt.Sprintf("failed to set sibling pointers on content_data=%s: %v", childID, updateErr))
		}
	}
}
