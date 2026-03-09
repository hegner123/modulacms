package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/model"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/transform"
	"github.com/hegner123/modulacms/internal/tree/core"
	"github.com/hegner123/modulacms/internal/utility"
)

// SlugHandler dispatches slug-based content delivery requests.
func SlugHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetSlugContent(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetSlugContent serves published snapshot content for public delivery,
// with a preview mode fallback that serves live draft data for authenticated users.
func apiGetSlugContent(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	// Check for preview mode.
	if r.URL.Query().Get("preview") == "true" {
		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "preview mode requires authentication", http.StatusForbidden)
			return
		}
		w.Header().Set("X-Robots-Tag", "noindex")
		apiGetSlugContentLive(w, r, svc)
		return
	}

	// Normal public delivery: serve from published snapshot.
	apiGetSlugContentPublished(w, r, svc)
}

// apiGetSlugContentPublished serves content from published snapshots.
// It looks up the route, finds the root content data, retrieves the published
// snapshot, deserializes it, and builds the tree for response.
// When i18n is enabled, resolves locale from ?locale= param, Accept-Language
// header, or default locale. ?locale=* returns lightweight metadata only.
func apiGetSlugContentPublished(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()
	c, cfgErr := svc.Config()
	if cfgErr != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}

	slug := strings.TrimPrefix(r.URL.Path, "/api/v1/content")
	if slug == "" {
		slug = "/"
	}

	// 1. Look up route by slug.
	route, err := d.GetRouteID(slug)
	if err != nil {
		utility.DefaultLogger.Error("GetRouteID failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Find the root content data for this route (the one with no parent).
	nullableRoute := types.NullableRouteID{ID: *route, Valid: true}
	contentData, err := d.ListContentDataByRoute(nullableRoute)
	if err != nil {
		utility.DefaultLogger.Error("ListContentDataByRoute failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var rootContentDataID types.ContentID
	found := false
	for _, cd := range *contentData {
		if !cd.ParentID.Valid {
			rootContentDataID = cd.ContentDataID
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "no root content data found for route", http.StatusNotFound)
		return
	}

	// 2b. Handle ?locale=* — return locale availability metadata only.
	if r.URL.Query().Get("locale") == "*" {
		meta, metaErr := svc.Locales.BuildLocaleMetadata(r.Context(), rootContentDataID)
		if metaErr != nil {
			utility.DefaultLogger.Error("BuildLocaleMetadata failed", metaErr)
			http.Error(w, "failed to build locale metadata", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta)
		return
	}

	// 3. Resolve locale and get the published snapshot.
	locale := svc.Locales.ResolveLocale(r)

	var version *db.ContentVersion
	var resolvedLocale string
	if locale != "" {
		// i18n enabled: try requested locale, then fallback chain, then default.
		resolvedLocale, version, err = svc.Locales.WalkFallback(r.Context(), rootContentDataID, locale)
		if err != nil || version == nil {
			// Try default locale as last resort.
			defaultLocale := c.I18nDefaultLocale()
			if defaultLocale != locale {
				version, err = d.GetPublishedSnapshot(rootContentDataID, defaultLocale)
				if err == nil {
					resolvedLocale = defaultLocale
				}
			}
		}
		// If still no snapshot, try empty-locale (pre-i18n content).
		if version == nil {
			version, err = d.GetPublishedSnapshot(rootContentDataID, "")
			if err == nil {
				resolvedLocale = ""
			}
		}
	} else {
		// i18n disabled: use empty locale (original behavior).
		version, err = d.GetPublishedSnapshot(rootContentDataID, "")
		resolvedLocale = ""
	}

	if version == nil {
		http.Error(w, "content not published", http.StatusNotFound)
		return
	}

	// 4. Deserialize the snapshot JSON.
	var snapshot publishing.Snapshot
	if err := json.Unmarshal([]byte(version.Snapshot), &snapshot); err != nil {
		utility.DefaultLogger.Error("snapshot unmarshal failed", err)
		http.Error(w, "failed to read published content", http.StatusInternalServerError)
		return
	}

	// 5. Convert snapshot JSON types back to DB types for model.BuildTree.
	cdSlice, err := publishing.SnapshotContentDataToSlice(snapshot.ContentData)
	if err != nil {
		utility.DefaultLogger.Error("snapshot content data conversion failed", err)
		http.Error(w, "failed to process published content", http.StatusInternalServerError)
		return
	}

	dtSlice, err := publishing.SnapshotDatatypesToSlice(snapshot.Datatypes)
	if err != nil {
		utility.DefaultLogger.Error("snapshot datatypes conversion failed", err)
		http.Error(w, "failed to process published content", http.StatusInternalServerError)
		return
	}

	cfSlice, err := publishing.SnapshotContentFieldsToSlice(snapshot.ContentFields)
	if err != nil {
		utility.DefaultLogger.Error("snapshot content fields conversion failed", err)
		http.Error(w, "failed to process published content", http.StatusInternalServerError)
		return
	}

	fdSlice, err := publishing.SnapshotFieldsToSlice(snapshot.Fields)
	if err != nil {
		utility.DefaultLogger.Error("snapshot fields conversion failed", err)
		http.Error(w, "failed to process published content", http.StatusInternalServerError)
		return
	}

	// 6. Build the tree from snapshot data.
	root, err := model.BuildTree(utility.DefaultLogger, cdSlice, dtSlice, cfSlice, fdSlice)
	if err != nil {
		utility.DefaultLogger.Error("BuildTree error from snapshot", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 7. Compose referenced subtrees using published snapshots.
	// Pass locale through so composed references use the same locale.
	if root.CoreRoot != nil {
		fetcher := core.NewCachedFetcher(&SnapshotTreeFetcher{Driver: d, Locale: resolvedLocale})
		composeErr := core.ComposeTrees(r.Context(), root.CoreRoot, fetcher, core.ComposeOptions{
			MaxDepth:       c.CompositionMaxDepth(),
			MaxConcurrency: 10,
		})
		if composeErr != nil {
			utility.DefaultLogger.Warn("snapshot composition error", composeErr)
		}
		root.RebuildFromCore()
	}

	// 8. Apply format/transform the same way as the live flow.
	applyFormatAndTransform(w, r, *c, d, root)
}

// apiGetSlugContentLive serves live draft content directly from the database.
// This is the original content delivery path, now used only for preview mode.
// When i18n is enabled, resolves locale from the request and filters content
// fields to the requested locale + non-translatable fields (locale="").
func apiGetSlugContentLive(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()
	c, cfgErr := svc.Config()
	if cfgErr != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}

	slug := strings.TrimPrefix(r.URL.Path, "/api/v1/content")
	if slug == "" {
		slug = "/"
	}

	route, err := d.GetRouteID(slug)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	nullableRoute := types.NullableRouteID{ID: *route, Valid: true}
	contentData, err := d.ListContentDataByRoute(nullableRoute)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dataSlice := *contentData

	// Fetch datatype definitions for each content data node.
	dt := []db.Datatypes{}
	for _, da := range dataSlice {
		if !da.DatatypeID.Valid {
			continue
		}
		datatype, dtErr := d.GetDatatype(da.DatatypeID.ID)
		if dtErr != nil {
			utility.DefaultLogger.Error("", dtErr)
			http.Error(w, dtErr.Error(), http.StatusInternalServerError)
			return
		}
		dt = append(dt, *datatype)
	}

	// Resolve locale for field filtering. When i18n is enabled, this returns
	// the target locale; when disabled, returns "" (fetch all fields).
	locale := svc.Locales.ResolveLocale(r)

	// Fetch content field values for this route, filtered by locale when i18n is enabled.
	var contentFields *[]db.ContentFields
	if locale != "" {
		// i18n enabled: fetch only locale-specific + non-translatable (locale="") fields.
		contentFields, err = d.ListContentFieldsByRouteAndLocale(nullableRoute, locale)
	} else {
		// i18n disabled: fetch all fields (original behavior).
		contentFields, err = d.ListContentFieldsByRoute(nullableRoute)
	}
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build parallel slices of content fields and field definitions,
	// starting with fields that already have content values.
	var allCF []db.ContentFields
	var allFD []db.Fields
	for _, cf := range *contentFields {
		if !cf.FieldID.Valid {
			continue
		}
		field, fErr := d.GetField(cf.FieldID.ID)
		if fErr != nil {
			utility.DefaultLogger.Error("", fErr)
			http.Error(w, fErr.Error(), http.StatusInternalServerError)
			return
		}
		allCF = append(allCF, cf)
		allFD = append(allFD, *field)
	}

	// Track which (content_data_id, field_id) pairs already have values.
	type fieldKey struct{ contentDataID, fieldID string }
	existing := make(map[fieldKey]bool)
	for _, cf := range allCF {
		if cf.ContentDataID.Valid && cf.FieldID.Valid {
			existing[fieldKey{cf.ContentDataID.ID.String(), cf.FieldID.ID.String()}] = true
		}
	}

	// For each content data node, look up all schema-defined fields for its
	// datatype and add empty stubs for any that don't have content values.
	for _, da := range dataSlice {
		if !da.DatatypeID.Valid {
			continue
		}
		dtID := types.NullableDatatypeID{ID: da.DatatypeID.ID, Valid: true}
		schemaFields, sfErr := d.ListFieldsByDatatypeID(dtID)
		if sfErr != nil {
			utility.DefaultLogger.Error("", sfErr)
			http.Error(w, sfErr.Error(), http.StatusInternalServerError)
			return
		}
		if schemaFields == nil {
			continue
		}
		for _, sf := range *schemaFields {
			key := fieldKey{da.ContentDataID.String(), sf.FieldID.String()}
			if existing[key] {
				continue
			}
			stub := db.ContentFields{
				ContentDataID: types.NullableContentID{ID: da.ContentDataID, Valid: true},
				FieldID:       types.NullableFieldID{ID: sf.FieldID, Valid: true},
				RouteID:       da.RouteID,
			}
			allCF = append(allCF, stub)
			allFD = append(allFD, sf)
		}
	}

	// Filter fields by the authenticated user's role for preview mode.
	// The caller (apiGetSlugContent) already verified the user is authenticated.
	user := middleware.AuthenticatedUser(r.Context())
	isAdmin := middleware.ContextIsAdmin(r.Context())
	roleID := ""
	if user != nil {
		roleID = user.Role
	}
	filteredFD := db.FilterFieldsByRole(allFD, roleID, isAdmin)
	if len(filteredFD) != len(allFD) {
		// Rebuild allCF to match the filtered allFD by index.
		accessibleFieldIDs := make(map[string]bool, len(filteredFD))
		for _, fd := range filteredFD {
			accessibleFieldIDs[fd.FieldID.String()] = true
		}
		var filteredCF []db.ContentFields
		var filteredFD2 []db.Fields
		for i, fd := range allFD {
			if accessibleFieldIDs[fd.FieldID.String()] {
				filteredCF = append(filteredCF, allCF[i])
				filteredFD2 = append(filteredFD2, fd)
			}
		}
		allCF = filteredCF
		allFD = filteredFD2
	}

	root, err := model.BuildTree(utility.DefaultLogger, dataSlice, dt, allCF, allFD)
	if err != nil {
		utility.DefaultLogger.Error("BuildTree error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Compose referenced subtrees using live data (preview mode sees live refs).
	if root.CoreRoot != nil {
		fetcher := core.NewCachedFetcher(&core.DbTreeFetcher{Driver: d})
		composeErr := core.ComposeTrees(r.Context(), root.CoreRoot, fetcher, core.ComposeOptions{
			MaxDepth:       c.CompositionMaxDepth(),
			MaxConcurrency: 10,
		})
		if composeErr != nil {
			utility.DefaultLogger.Warn("composition error", composeErr)
		}
		root.RebuildFromCore()
	}

	applyFormatAndTransform(w, r, *c, d, root)
}

// applyFormatAndTransform applies the output format and writes the transformed
// response. Shared by both the published and live delivery paths.
func applyFormatAndTransform(w http.ResponseWriter, r *http.Request, c config.Config, d db.DbDriver, root model.Root) {
	// Allow format override via query parameter.
	format := c.Output_Format
	if queryFormat := r.URL.Query().Get("format"); queryFormat != "" {
		if config.IsValidOutputFormat(queryFormat) {
			format = config.OutputFormat(queryFormat)
		} else {
			http.Error(w, "Invalid format parameter. Valid options: contentful, sanity, strapi, wordpress, clean, raw", http.StatusBadRequest)
			return
		}
	}

	// Create transform config and write response.
	transformCfg := transform.NewTransformConfig(
		format,
		c.Client_Site,
		c.Space_ID,
		d,
	)

	if err := transformCfg.TransformAndWrite(w, root); err != nil {
		utility.DefaultLogger.Error("Transform error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

///////////////////////////////
// SNAPSHOT TREE FETCHER
///////////////////////////////

// SnapshotTreeFetcher implements core.TreeFetcher by resolving content
// references via published snapshots instead of live database data.
// This ensures that public delivery only shows published content for
// referenced subtrees. Locale propagates to composed references so that
// all subtrees resolve from the same locale.
type SnapshotTreeFetcher struct {
	Driver db.DbDriver
	Locale string
}

// FetchAndBuildTree retrieves the published snapshot for the given content
// data ID, deserializes it, and builds the tree. If no published snapshot
// exists for the reference, it returns nil gracefully (the composition
// layer will produce a _system_log node for the missing reference).
func (f *SnapshotTreeFetcher) FetchAndBuildTree(ctx context.Context, id types.ContentID) (*core.Root, error) {
	version, err := f.Driver.GetPublishedSnapshot(id, f.Locale)
	if err != nil {
		return nil, fmt.Errorf("no published snapshot for %s: %w", id, err)
	}

	var snapshot publishing.Snapshot
	if err := json.Unmarshal([]byte(version.Snapshot), &snapshot); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot for %s: %w", id, err)
	}

	cd, err := publishing.SnapshotContentDataToSlice(snapshot.ContentData)
	if err != nil {
		return nil, fmt.Errorf("convert snapshot content data for %s: %w", id, err)
	}

	dt, err := publishing.SnapshotDatatypesToSlice(snapshot.Datatypes)
	if err != nil {
		return nil, fmt.Errorf("convert snapshot datatypes for %s: %w", id, err)
	}

	cf, err := publishing.SnapshotContentFieldsToSlice(snapshot.ContentFields)
	if err != nil {
		return nil, fmt.Errorf("convert snapshot content fields for %s: %w", id, err)
	}

	df, err := publishing.SnapshotFieldsToSlice(snapshot.Fields)
	if err != nil {
		return nil, fmt.Errorf("convert snapshot fields for %s: %w", id, err)
	}

	// Set the root's datatype type to _nested_root (same as DbTreeFetcher).
	for i, c := range cd {
		if c.ContentDataID == id {
			dt[i].Type = string(types.DatatypeTypeNestedRoot)
			break
		}
	}

	root, _, err := core.BuildTree(cd, dt, cf, df)
	return root, err
}
