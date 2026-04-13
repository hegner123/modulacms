package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
)

// ---------------------------------------------------------------------------
// MediaBackend
// ---------------------------------------------------------------------------

type svcMediaBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcMediaBackend) ListMedia(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	items, total, err := b.svc.Media.ListMediaPaginated(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	resp := db.PaginatedResponse[mcpMediaResponse]{
		Limit:  limit,
		Offset: offset,
	}
	if items != nil {
		resp.Data = toMCPMediaList(*items)
	}
	if total != nil {
		resp.Total = *total
	}
	return json.Marshal(resp)
}

func (b *svcMediaBackend) GetMedia(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Media.GetMedia(ctx, types.MediaID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(toMCPMediaResponse(*result))
}

func (b *svcMediaBackend) UpdateMedia(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.UpdateMediaMetadataParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update media params: %w", err)
	}
	result, err := b.svc.Media.UpdateMediaMetadata(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcMediaBackend) DeleteMedia(ctx context.Context, id string) error {
	return b.svc.Media.DeleteMedia(ctx, b.ac, types.MediaID(id))
}

func (b *svcMediaBackend) UploadMedia(ctx context.Context, reader io.Reader, filename string) (json.RawMessage, error) {
	// The MediaService.Upload requires multipart.File and *multipart.FileHeader,
	// which are HTTP-specific types. In direct mode we cannot easily construct
	// these from an io.Reader. Return an unsupported error directing callers
	// to use the REST API for media upload.
	return nil, fmt.Errorf("media upload is not supported in direct mode; use the REST API")
}

func (b *svcMediaBackend) MediaHealth(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Media.MediaHealth(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcMediaBackend) MediaCleanup(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Media.MediaCleanup(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcMediaBackend) MediaCleanupCheck(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Media.MediaHealth(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcMediaBackend) ListMediaDimensions(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Media.ListMediaDimensions(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcMediaBackend) GetMediaDimension(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Media.GetMediaDimension(ctx, id)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcMediaBackend) CreateMediaDimension(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateMediaDimensionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create media dimension params: %w", err)
	}
	result, err := b.svc.Media.CreateMediaDimension(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcMediaBackend) UpdateMediaDimension(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateMediaDimensionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update media dimension params: %w", err)
	}
	result, err := b.svc.Media.UpdateMediaDimension(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcMediaBackend) DeleteMediaDimension(ctx context.Context, id string) error {
	return b.svc.Media.DeleteMediaDimension(ctx, b.ac, id)
}

func (b *svcMediaBackend) DownloadMedia(ctx context.Context, id string) (json.RawMessage, error) {
	mid := types.MediaID(id)
	if err := mid.Validate(); err != nil {
		return nil, fmt.Errorf("invalid media id: %w", err)
	}
	record, err := b.svc.Media.GetMedia(ctx, mid)
	if err != nil {
		return nil, err
	}
	// In direct mode, return the stored public URL (pre-signed S3 URLs require
	// HTTP request context). SDK mode generates a true download redirect URL.
	return json.Marshal(map[string]string{"url": string(record.URL)})
}

func (b *svcMediaBackend) GetMediaFull(ctx context.Context) (json.RawMessage, error) {
	views, err := b.svc.Media.ListMediaFull(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(views)
}

func (b *svcMediaBackend) GetMediaReferences(ctx context.Context, id string) (json.RawMessage, error) {
	d := b.svc.Driver()

	mid := types.MediaID(id)
	if err := mid.Validate(); err != nil {
		return nil, fmt.Errorf("invalid media id: %w", err)
	}

	mediaRecord, err := d.GetMedia(mid)
	if err != nil {
		return nil, &service.NotFoundError{Resource: "media", ID: id}
	}

	searchTerms := []string{string(mid)}
	if string(mediaRecord.URL) != "" {
		searchTerms = append(searchTerms, string(mediaRecord.URL))
	}

	allFields, err := d.ListContentFields()
	if err != nil {
		return nil, fmt.Errorf("list content fields for reference scan: %w", err)
	}

	type refInfo struct {
		ContentFieldID types.ContentFieldID    `json:"content_field_id"`
		ContentDataID  types.NullableContentID `json:"content_data_id"`
		FieldID        types.NullableFieldID   `json:"field_id"`
	}

	refs := make([]refInfo, 0)
	if allFields != nil {
		for _, cf := range *allFields {
			if cf.FieldValue == "" {
				continue
			}
			for _, term := range searchTerms {
				if strings.Contains(cf.FieldValue, term) {
					refs = append(refs, refInfo{
						ContentFieldID: cf.ContentFieldID,
						ContentDataID:  cf.ContentDataID,
						FieldID:        cf.FieldID,
					})
					break
				}
			}
		}
	}

	return json.Marshal(map[string]any{
		"media_id":        mid,
		"references":      refs,
		"reference_count": len(refs),
	})
}

func (b *svcMediaBackend) ReprocessMedia(ctx context.Context) (json.RawMessage, error) {
	started := b.svc.Media.TriggerReprocess()
	if started {
		return json.Marshal(map[string]any{"reprocess_started": true, "message": "bulk reprocess started"})
	}
	return json.Marshal(map[string]any{"reprocess_started": false, "message": "reprocess already running, restart queued"})
}

// ---------------------------------------------------------------------------
// MediaFolderBackend
// ---------------------------------------------------------------------------

type svcMediaFolderBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcMediaFolderBackend) ListMediaFolders(ctx context.Context, parentID string) (json.RawMessage, error) {
	d := b.svc.Driver()
	if parentID != "" {
		pid := types.MediaFolderID(parentID)
		if err := pid.Validate(); err != nil {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: fmt.Sprintf("invalid: %v", err)}}}
		}
		folders, err := d.ListMediaFoldersByParent(pid)
		if err != nil {
			return nil, err
		}
		return json.Marshal(folders)
	}
	folders, err := d.ListMediaFoldersAtRoot()
	if err != nil {
		return nil, err
	}
	return json.Marshal(folders)
}

func (b *svcMediaFolderBackend) GetMediaFolder(ctx context.Context, id string) (json.RawMessage, error) {
	d := b.svc.Driver()
	fid := types.MediaFolderID(id)
	if err := fid.Validate(); err != nil {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}
	folder, err := d.GetMediaFolder(fid)
	if err != nil {
		return nil, &service.NotFoundError{Resource: "media_folder", ID: id}
	}
	return json.Marshal(folder)
}

func (b *svcMediaFolderBackend) CreateMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	d := b.svc.Driver()

	var p struct {
		Name     string  `json:"name"`
		ParentID *string `json:"parent_id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create media folder params: %w", err)
	}
	if p.Name == "" {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "name", Message: "required"}}}
	}

	var parentID types.NullableMediaFolderID
	if p.ParentID != nil && *p.ParentID != "" {
		pid := types.MediaFolderID(*p.ParentID)
		if err := pid.Validate(); err != nil {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: fmt.Sprintf("invalid: %v", err)}}}
		}
		if _, err := d.GetMediaFolder(pid); err != nil {
			return nil, &service.NotFoundError{Resource: "parent_folder", ID: *p.ParentID}
		}
		parentID = types.NullableMediaFolderID{ID: pid, Valid: true}

		breadcrumb, err := d.GetMediaFolderBreadcrumb(pid)
		if err != nil {
			return nil, fmt.Errorf("check folder depth: %w", err)
		}
		if len(breadcrumb)+1 > 10 {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: "creating this folder would exceed maximum folder depth of 10"}}}
		}
	}

	if err := d.ValidateMediaFolderName(p.Name, parentID); err != nil {
		return nil, &service.ConflictError{Resource: "media_folder", Detail: err.Error()}
	}

	now := types.NewTimestamp(time.Now().UTC())
	folder, err := d.CreateMediaFolder(ctx, b.ac, db.CreateMediaFolderParams{
		Name:         p.Name,
		ParentID:     parentID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(folder)
}

func (b *svcMediaFolderBackend) UpdateMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	d := b.svc.Driver()

	var p struct {
		FolderID string  `json:"folder_id"`
		Name     *string `json:"name"`
		ParentID *string `json:"parent_id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update media folder params: %w", err)
	}

	fid := types.MediaFolderID(p.FolderID)
	if err := fid.Validate(); err != nil {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "folder_id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}

	existing, err := d.GetMediaFolder(fid)
	if err != nil {
		return nil, &service.NotFoundError{Resource: "media_folder", ID: p.FolderID}
	}

	name := existing.Name
	parentID := existing.ParentID

	if p.Name != nil {
		if *p.Name == "" {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "name", Message: "cannot be empty"}}}
		}
		name = *p.Name
	}

	parentChanged := false
	if p.ParentID != nil {
		parentChanged = true
		if *p.ParentID == "" {
			parentID = types.NullableMediaFolderID{}
		} else {
			pid := types.MediaFolderID(*p.ParentID)
			if err := pid.Validate(); err != nil {
				return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: fmt.Sprintf("invalid: %v", err)}}}
			}
			parentID = types.NullableMediaFolderID{ID: pid, Valid: true}
		}
	}

	if parentChanged {
		if err := d.ValidateMediaFolderMove(fid, parentID); err != nil {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: err.Error()}}}
		}
	}

	nameChanged := p.Name != nil && name != existing.Name
	if nameChanged || parentChanged {
		if err := d.ValidateMediaFolderName(name, parentID); err != nil {
			return nil, &service.ConflictError{Resource: "media_folder", Detail: err.Error()}
		}
	}

	_, err = d.UpdateMediaFolder(ctx, b.ac, db.UpdateMediaFolderParams{
		FolderID:     fid,
		Name:         name,
		ParentID:     parentID,
		DateModified: types.NewTimestamp(time.Now().UTC()),
	})
	if err != nil {
		return nil, err
	}

	updated, err := d.GetMediaFolder(fid)
	if err != nil {
		return nil, err
	}
	return json.Marshal(updated)
}

func (b *svcMediaFolderBackend) DeleteMediaFolder(ctx context.Context, id string) (json.RawMessage, error) {
	d := b.svc.Driver()

	fid := types.MediaFolderID(id)
	if err := fid.Validate(); err != nil {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}

	if _, err := d.GetMediaFolder(fid); err != nil {
		return nil, &service.NotFoundError{Resource: "media_folder", ID: id}
	}

	children, err := d.ListMediaFoldersByParent(fid)
	if err != nil {
		return nil, err
	}
	childCount := 0
	if children != nil {
		childCount = len(*children)
	}

	folderNullable := types.NullableMediaFolderID{ID: fid, Valid: true}
	mediaCount, err := d.CountMediaByFolder(folderNullable)
	if err != nil {
		return nil, err
	}

	mc := int64(0)
	if mediaCount != nil {
		mc = *mediaCount
	}

	if childCount > 0 || mc > 0 {
		return json.Marshal(map[string]any{
			"error":         fmt.Sprintf("cannot delete folder: contains %d child folder(s) and %d media item(s)", childCount, mc),
			"child_folders": childCount,
			"media_items":   mc,
		})
	}

	if err := d.DeleteMediaFolder(ctx, b.ac, fid); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{"status": "deleted"})
}

func (b *svcMediaFolderBackend) MoveMediaToFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	d := b.svc.Driver()

	var p struct {
		MediaIDs []string `json:"media_ids"`
		FolderID *string  `json:"folder_id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal move media params: %w", err)
	}

	if len(p.MediaIDs) == 0 {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "media_ids", Message: "required and cannot be empty"}}}
	}
	if len(p.MediaIDs) > 100 {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "media_ids", Message: "batch size cannot exceed 100 items"}}}
	}

	var folderID types.NullableMediaFolderID
	if p.FolderID != nil && *p.FolderID != "" {
		fid := types.MediaFolderID(*p.FolderID)
		if err := fid.Validate(); err != nil {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "folder_id", Message: fmt.Sprintf("invalid: %v", err)}}}
		}
		if _, err := d.GetMediaFolder(fid); err != nil {
			return nil, &service.NotFoundError{Resource: "media_folder", ID: *p.FolderID}
		}
		folderID = types.NullableMediaFolderID{ID: fid, Valid: true}
	}

	now := types.NewTimestamp(time.Now().UTC())
	moved := 0
	for _, idStr := range p.MediaIDs {
		mid := types.MediaID(idStr)
		if err := mid.Validate(); err != nil {
			continue
		}
		err := d.MoveMediaToFolder(ctx, b.ac, db.MoveMediaToFolderParams{
			FolderID:     folderID,
			DateModified: now,
			MediaID:      mid,
		})
		if err != nil {
			continue
		}
		moved++
	}

	return json.Marshal(map[string]int{"moved": moved})
}

func (b *svcMediaFolderBackend) GetMediaFolderTree(ctx context.Context) (json.RawMessage, error) {
	d := b.svc.Driver()
	allFolders, err := d.ListMediaFolders()
	if err != nil {
		return nil, err
	}
	tree := buildMediaFolderTree(allFolders)
	return json.Marshal(tree)
}

func (b *svcMediaFolderBackend) ListMediaInFolder(ctx context.Context, folderID string, limit, offset int64) (json.RawMessage, error) {
	d := b.svc.Driver()
	fid := types.MediaFolderID(folderID)
	if err := fid.Validate(); err != nil {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "folder_id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}
	if _, err := d.GetMediaFolder(fid); err != nil {
		return nil, &service.NotFoundError{Resource: "media_folder", ID: folderID}
	}
	folderNullable := types.NullableMediaFolderID{ID: fid, Valid: true}
	items, err := d.ListMediaByFolderPaginated(db.ListMediaByFolderPaginatedParams{
		FolderID: folderNullable,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}
	total, err := d.CountMediaByFolder(folderNullable)
	if err != nil {
		return nil, err
	}
	resp := db.PaginatedResponse[db.Media]{
		Limit:  limit,
		Offset: offset,
	}
	if items != nil {
		resp.Data = *items
	}
	if total != nil {
		resp.Total = *total
	}
	return json.Marshal(resp)
}

// ---------------------------------------------------------------------------
// AdminMediaBackend
// ---------------------------------------------------------------------------

type svcAdminMediaBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcAdminMediaBackend) ListAdminMedia(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
	d := b.svc.Driver()
	items, err := d.ListAdminMediaPaginated(db.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	total, err := d.CountAdminMedia()
	if err != nil {
		return nil, err
	}
	resp := db.PaginatedResponse[mcpAdminMediaResponse]{
		Limit:  limit,
		Offset: offset,
	}
	if items != nil {
		resp.Data = toMCPAdminMediaList(*items)
	}
	if total != nil {
		resp.Total = *total
	}
	return json.Marshal(resp)
}

func (b *svcAdminMediaBackend) GetAdminMedia(ctx context.Context, id string) (json.RawMessage, error) {
	d := b.svc.Driver()
	mid := types.AdminMediaID(id)
	if err := mid.Validate(); err != nil {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}
	result, err := d.GetAdminMedia(mid)
	if err != nil {
		return nil, &service.NotFoundError{Resource: "admin_media", ID: id}
	}
	return json.Marshal(toMCPAdminMediaResponse(*result))
}

func (b *svcAdminMediaBackend) UpdateAdminMedia(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	d := b.svc.Driver()

	var p struct {
		AdminMediaID string   `json:"admin_media_id"`
		Name         *string  `json:"name"`
		DisplayName  *string  `json:"display_name"`
		Alt          *string  `json:"alt"`
		Caption      *string  `json:"caption"`
		Description  *string  `json:"description"`
		Class        *string  `json:"class"`
		FocalX       *float64 `json:"focal_x"`
		FocalY       *float64 `json:"focal_y"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin media params: %w", err)
	}

	mid := types.AdminMediaID(p.AdminMediaID)
	if err := mid.Validate(); err != nil {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "admin_media_id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}

	existing, err := d.GetAdminMedia(mid)
	if err != nil {
		return nil, &service.NotFoundError{Resource: "admin_media", ID: p.AdminMediaID}
	}

	updateParams := db.UpdateAdminMediaParams{
		AdminMediaID: mid,
		Name:         existing.Name,
		DisplayName:  existing.DisplayName,
		Alt:          existing.Alt,
		Caption:      existing.Caption,
		Description:  existing.Description,
		Class:        existing.Class,
		Mimetype:     existing.Mimetype,
		Dimensions:   existing.Dimensions,
		URL:          existing.URL,
		Srcset:       existing.Srcset,
		FocalX:       existing.FocalX,
		FocalY:       existing.FocalY,
		AuthorID:     existing.AuthorID,
		FolderID:     existing.FolderID,
		DateCreated:  existing.DateCreated,
		DateModified: types.NewTimestamp(time.Now().UTC()),
	}

	if p.Name != nil {
		updateParams.Name = db.NewNullString(*p.Name)
	}
	if p.DisplayName != nil {
		updateParams.DisplayName = db.NewNullString(*p.DisplayName)
	}
	if p.Alt != nil {
		updateParams.Alt = db.NewNullString(*p.Alt)
	}
	if p.Caption != nil {
		updateParams.Caption = db.NewNullString(*p.Caption)
	}
	if p.Description != nil {
		updateParams.Description = db.NewNullString(*p.Description)
	}
	if p.Class != nil {
		updateParams.Class = db.NewNullString(*p.Class)
	}
	if p.FocalX != nil {
		updateParams.FocalX = types.NullableFloat64{Float64: *p.FocalX, Valid: true}
	}
	if p.FocalY != nil {
		updateParams.FocalY = types.NullableFloat64{Float64: *p.FocalY, Valid: true}
	}

	if _, err := d.UpdateAdminMedia(ctx, b.ac, updateParams); err != nil {
		return nil, err
	}

	updated, err := d.GetAdminMedia(mid)
	if err != nil {
		return nil, err
	}
	return json.Marshal(toMCPAdminMediaResponse(*updated))
}

func (b *svcAdminMediaBackend) DeleteAdminMedia(ctx context.Context, id string) error {
	d := b.svc.Driver()
	mid := types.AdminMediaID(id)
	if err := mid.Validate(); err != nil {
		return &service.ValidationError{Errors: []service.FieldError{{Field: "id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}
	return d.DeleteAdminMedia(ctx, b.ac, mid)
}

func (b *svcAdminMediaBackend) UploadAdminMedia(ctx context.Context, reader io.Reader, filename string) (json.RawMessage, error) {
	// The MediaService.Upload requires multipart.File and *multipart.FileHeader,
	// which are HTTP-specific types. In direct mode we cannot easily construct
	// these from an io.Reader. Return an unsupported error directing callers
	// to use the REST API for admin media upload.
	return nil, fmt.Errorf("admin media upload is not supported in direct mode; use the REST API")
}

func (b *svcAdminMediaBackend) ListMediaDimensions(ctx context.Context) (json.RawMessage, error) {
	// Dimensions are shared between public and admin media.
	result, err := b.svc.Media.ListMediaDimensions(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// ---------------------------------------------------------------------------
// AdminMediaFolderBackend
// ---------------------------------------------------------------------------

type svcAdminMediaFolderBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcAdminMediaFolderBackend) ListAdminMediaFolders(ctx context.Context, parentID string) (json.RawMessage, error) {
	d := b.svc.Driver()
	if parentID != "" {
		pid := types.AdminMediaFolderID(parentID)
		if err := pid.Validate(); err != nil {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: fmt.Sprintf("invalid: %v", err)}}}
		}
		folders, err := d.ListAdminMediaFoldersByParent(pid)
		if err != nil {
			return nil, err
		}
		return json.Marshal(folders)
	}
	folders, err := d.ListAdminMediaFoldersAtRoot()
	if err != nil {
		return nil, err
	}
	return json.Marshal(folders)
}

func (b *svcAdminMediaFolderBackend) GetAdminMediaFolder(ctx context.Context, id string) (json.RawMessage, error) {
	d := b.svc.Driver()
	fid := types.AdminMediaFolderID(id)
	if err := fid.Validate(); err != nil {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}
	folder, err := d.GetAdminMediaFolder(fid)
	if err != nil {
		return nil, &service.NotFoundError{Resource: "admin_media_folder", ID: id}
	}
	return json.Marshal(folder)
}

func (b *svcAdminMediaFolderBackend) CreateAdminMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	d := b.svc.Driver()

	var p struct {
		Name     string  `json:"name"`
		ParentID *string `json:"parent_id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin media folder params: %w", err)
	}
	if p.Name == "" {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "name", Message: "required"}}}
	}

	var parentID types.NullableAdminMediaFolderID
	if p.ParentID != nil && *p.ParentID != "" {
		pid := types.AdminMediaFolderID(*p.ParentID)
		if err := pid.Validate(); err != nil {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: fmt.Sprintf("invalid: %v", err)}}}
		}
		if _, err := d.GetAdminMediaFolder(pid); err != nil {
			return nil, &service.NotFoundError{Resource: "parent_folder", ID: *p.ParentID}
		}
		parentID = types.NullableAdminMediaFolderID{ID: pid, Valid: true}

		breadcrumb, err := d.GetAdminMediaFolderBreadcrumb(pid)
		if err != nil {
			return nil, fmt.Errorf("check folder depth: %w", err)
		}
		if len(breadcrumb)+1 > 10 {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: "creating this folder would exceed maximum folder depth of 10"}}}
		}
	}

	if err := d.ValidateAdminMediaFolderName(p.Name, parentID); err != nil {
		return nil, &service.ConflictError{Resource: "admin_media_folder", Detail: err.Error()}
	}

	now := types.NewTimestamp(time.Now().UTC())
	folder, err := d.CreateAdminMediaFolder(ctx, b.ac, db.CreateAdminMediaFolderParams{
		Name:         p.Name,
		ParentID:     parentID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(folder)
}

func (b *svcAdminMediaFolderBackend) UpdateAdminMediaFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	d := b.svc.Driver()

	var p struct {
		FolderID string  `json:"folder_id"`
		Name     *string `json:"name"`
		ParentID *string `json:"parent_id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin media folder params: %w", err)
	}

	fid := types.AdminMediaFolderID(p.FolderID)
	if err := fid.Validate(); err != nil {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "folder_id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}

	existing, err := d.GetAdminMediaFolder(fid)
	if err != nil {
		return nil, &service.NotFoundError{Resource: "admin_media_folder", ID: p.FolderID}
	}

	name := existing.Name
	parentID := existing.ParentID

	if p.Name != nil {
		if *p.Name == "" {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "name", Message: "cannot be empty"}}}
		}
		name = *p.Name
	}

	parentChanged := false
	if p.ParentID != nil {
		parentChanged = true
		if *p.ParentID == "" {
			parentID = types.NullableAdminMediaFolderID{}
		} else {
			pid := types.AdminMediaFolderID(*p.ParentID)
			if err := pid.Validate(); err != nil {
				return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: fmt.Sprintf("invalid: %v", err)}}}
			}
			parentID = types.NullableAdminMediaFolderID{ID: pid, Valid: true}
		}
	}

	if parentChanged {
		if err := d.ValidateAdminMediaFolderMove(fid, parentID); err != nil {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "parent_id", Message: err.Error()}}}
		}
	}

	nameChanged := p.Name != nil && name != existing.Name
	if nameChanged || parentChanged {
		if err := d.ValidateAdminMediaFolderName(name, parentID); err != nil {
			return nil, &service.ConflictError{Resource: "admin_media_folder", Detail: err.Error()}
		}
	}

	_, err = d.UpdateAdminMediaFolder(ctx, b.ac, db.UpdateAdminMediaFolderParams{
		AdminFolderID: fid,
		Name:          name,
		ParentID:      parentID,
		DateModified:  types.NewTimestamp(time.Now().UTC()),
	})
	if err != nil {
		return nil, err
	}

	updated, err := d.GetAdminMediaFolder(fid)
	if err != nil {
		return nil, err
	}
	return json.Marshal(updated)
}

func (b *svcAdminMediaFolderBackend) DeleteAdminMediaFolder(ctx context.Context, id string) (json.RawMessage, error) {
	d := b.svc.Driver()

	fid := types.AdminMediaFolderID(id)
	if err := fid.Validate(); err != nil {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}

	if _, err := d.GetAdminMediaFolder(fid); err != nil {
		return nil, &service.NotFoundError{Resource: "admin_media_folder", ID: id}
	}

	children, err := d.ListAdminMediaFoldersByParent(fid)
	if err != nil {
		return nil, err
	}
	childCount := 0
	if children != nil {
		childCount = len(*children)
	}

	folderNullable := types.NullableAdminMediaFolderID{ID: fid, Valid: true}
	mediaCount, err := d.CountAdminMediaByFolder(folderNullable)
	if err != nil {
		return nil, err
	}

	mc := int64(0)
	if mediaCount != nil {
		mc = *mediaCount
	}

	if childCount > 0 || mc > 0 {
		return json.Marshal(map[string]any{
			"error":         fmt.Sprintf("cannot delete folder: contains %d child folder(s) and %d media item(s)", childCount, mc),
			"child_folders": childCount,
			"media_items":   mc,
		})
	}

	if err := d.DeleteAdminMediaFolder(ctx, b.ac, fid); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{"status": "deleted"})
}

func (b *svcAdminMediaFolderBackend) MoveAdminMediaToFolder(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	d := b.svc.Driver()

	var p struct {
		MediaIDs []string `json:"media_ids"`
		FolderID *string  `json:"folder_id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal move admin media params: %w", err)
	}

	if len(p.MediaIDs) == 0 {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "media_ids", Message: "required and cannot be empty"}}}
	}
	if len(p.MediaIDs) > 100 {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "media_ids", Message: "batch size cannot exceed 100 items"}}}
	}

	var folderID types.NullableAdminMediaFolderID
	if p.FolderID != nil && *p.FolderID != "" {
		fid := types.AdminMediaFolderID(*p.FolderID)
		if err := fid.Validate(); err != nil {
			return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "folder_id", Message: fmt.Sprintf("invalid: %v", err)}}}
		}
		if _, err := d.GetAdminMediaFolder(fid); err != nil {
			return nil, &service.NotFoundError{Resource: "admin_media_folder", ID: *p.FolderID}
		}
		folderID = types.NullableAdminMediaFolderID{ID: fid, Valid: true}
	}

	now := types.NewTimestamp(time.Now().UTC())
	moved := 0
	for _, idStr := range p.MediaIDs {
		mid := types.AdminMediaID(idStr)
		if err := mid.Validate(); err != nil {
			continue
		}
		err := d.MoveAdminMediaToFolder(ctx, b.ac, db.MoveAdminMediaToFolderParams{
			FolderID:     folderID,
			DateModified: now,
			AdminMediaID: mid,
		})
		if err != nil {
			continue
		}
		moved++
	}

	return json.Marshal(map[string]int{"moved": moved})
}

func (b *svcAdminMediaFolderBackend) AdminGetMediaFolderTree(ctx context.Context) (json.RawMessage, error) {
	d := b.svc.Driver()
	allFolders, err := d.ListAdminMediaFolders()
	if err != nil {
		return nil, err
	}
	tree := buildAdminMediaFolderTree(allFolders)
	return json.Marshal(tree)
}

func (b *svcAdminMediaFolderBackend) AdminListMediaInFolder(ctx context.Context, folderID string, limit, offset int64) (json.RawMessage, error) {
	d := b.svc.Driver()
	fid := types.AdminMediaFolderID(folderID)
	if err := fid.Validate(); err != nil {
		return nil, &service.ValidationError{Errors: []service.FieldError{{Field: "folder_id", Message: fmt.Sprintf("invalid: %v", err)}}}
	}
	if _, err := d.GetAdminMediaFolder(fid); err != nil {
		return nil, &service.NotFoundError{Resource: "admin_media_folder", ID: folderID}
	}
	folderNullable := types.NullableAdminMediaFolderID{ID: fid, Valid: true}
	items, err := d.ListAdminMediaByFolderPaginated(db.ListAdminMediaByFolderPaginatedParams{
		FolderID: folderNullable,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}
	total, err := d.CountAdminMediaByFolder(folderNullable)
	if err != nil {
		return nil, err
	}
	resp := db.PaginatedResponse[db.AdminMedia]{
		Limit:  limit,
		Offset: offset,
	}
	if items != nil {
		resp.Data = *items
	}
	if total != nil {
		resp.Total = *total
	}
	return json.Marshal(resp)
}
