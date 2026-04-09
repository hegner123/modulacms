package service

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/model"
	"github.com/hegner123/modulacms/internal/transform"
	"github.com/hegner123/modulacms/internal/utility"
)

// ImportService orchestrates multi-format content import.
type ImportService struct {
	driver db.DbDriver
	mgr    *config.Manager
}

// NewImportService creates an ImportService.
func NewImportService(driver db.DbDriver, mgr *config.Manager) *ImportService {
	return &ImportService{driver: driver, mgr: mgr}
}

// ImportContentInput holds the parameters for a content import operation.
type ImportContentInput struct {
	Format  config.OutputFormat
	Body    []byte // raw JSON body (already read)
	RouteID types.NullableRouteID
}

// ImportResult represents the result of an import operation.
type ImportResult struct {
	Success          bool     `json:"success"`
	DatatypesCreated int      `json:"datatypes_created"`
	FieldsCreated    int      `json:"fields_created"`
	ContentCreated   int      `json:"content_created"`
	Errors           []string `json:"errors,omitempty"`
	Message          string   `json:"message"`
}

// ImportContent parses the given body using the specified CMS format transformer,
// then imports the resulting content tree into the database.
func (s *ImportService) ImportContent(ctx context.Context, ac audited.AuditContext, input ImportContentInput) (*ImportResult, error) {
	cfg, err := s.mgr.Config()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	// Create transformer for the specified format
	transformCfg := transform.NewTransformConfig(
		transform.OutputFormat(input.Format),
		cfg.Client_Site,
		cfg.Space_ID,
		s.driver,
	)

	transformer, err := transformCfg.GetTransformer()
	if err != nil {
		utility.DefaultLogger.Error("failed to get transformer", err)
		return nil, fmt.Errorf("get transformer: %w", err)
	}

	// Parse CMS format to Modula
	root, err := transformer.Parse(input.Body)
	if err != nil {
		utility.DefaultLogger.Error("failed to parse input", err)
		return nil, fmt.Errorf("parse input: %w", err)
	}

	// Import to database
	result, err := s.importRootToDatabase(ctx, ac, root, input.RouteID)
	if err != nil {
		utility.DefaultLogger.Error("failed to import to database", err)
		return nil, fmt.Errorf("import to database: %w", err)
	}

	return result, nil
}

// importContext tracks state during recursive import.
type importContext struct {
	ctx           context.Context
	ac            audited.AuditContext
	driver        db.DbDriver
	routeID       types.NullableRouteID
	authorID      types.UserID
	datatypeCache map[string]types.DatatypeID // "label|type" -> existing DatatypeID
	result        *ImportResult
}

// importRootToDatabase imports a parsed Root structure to the database.
func (s *ImportService) importRootToDatabase(reqCtx context.Context, ac audited.AuditContext, root model.Root, routeID types.NullableRouteID) (*ImportResult, error) {
	result := &ImportResult{
		Success: false,
		Errors:  []string{},
	}

	if root.Node == nil {
		result.Errors = append(result.Errors, "no content to import")
		return result, nil
	}

	ctx := &importContext{
		ctx:           reqCtx,
		ac:            ac,
		driver:        s.driver,
		routeID:       routeID,
		authorID:      ac.UserID,
		datatypeCache: make(map[string]types.DatatypeID),
		result:        result,
	}

	ctx.importNode(root.Node, types.NullableContentID{Valid: false})

	result.Success = len(result.Errors) == 0
	if result.Success {
		result.Message = fmt.Sprintf("import complete: %d datatypes, %d fields, %d content nodes created",
			result.DatatypesCreated, result.FieldsCreated, result.ContentCreated)
	} else {
		result.Message = fmt.Sprintf("import completed with %d errors", len(result.Errors))
	}

	utility.DefaultLogger.Info("import completed",
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
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
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
		ParentID:     types.NullableDatatypeID{Valid: false},
		Name:         node.Datatype.Info.Name,
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

// createFieldAndContentField creates a field definition with parent_id set to the
// datatype, then creates a content_field linking it to the content_data.
func (ctx *importContext) createFieldAndContentField(field model.Field, contentDataID types.ContentID, datatypeID types.DatatypeID) {
	now := types.TimestampNow()

	// Create the field definition
	createdField, fieldErr := ctx.driver.CreateField(ctx.ctx, ctx.ac, db.CreateFieldParams{
		ParentID: types.NullableDatatypeID{
			ID:    datatypeID,
			Valid: true,
		},
		Name:         field.Info.Name,
		Label:        field.Info.Label,
		Data:         "",
		ValidationID: types.NullableValidationID{},
		UIConfig:     types.EmptyJSON,
		Type:         types.FieldType(field.Info.Type),
		AuthorID:     types.NullableUserID{ID: ctx.authorID, Valid: !ctx.authorID.IsZero()},
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
		FirstChildID:  types.NullableContentID{ID: childIDs[0], Valid: true},
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
		prevSibling := types.NullableContentID{}
		if i > 0 {
			prevSibling = types.NullableContentID{ID: childIDs[i-1], Valid: true}
		}

		nextSibling := types.NullableContentID{}
		if i < len(childIDs)-1 {
			nextSibling = types.NullableContentID{ID: childIDs[i+1], Valid: true}
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
