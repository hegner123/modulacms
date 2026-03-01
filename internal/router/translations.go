package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// translationRequest is the JSON body for creating a translation.
type translationRequest struct {
	Locale string `json:"locale"`
}

// translationResponse is the JSON body returned after creating a translation.
type translationResponse struct {
	Locale        string `json:"locale"`
	FieldsCreated int    `json:"fields_created"`
}

// TranslationHandler dispatches translation operations for content data.
func TranslationHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodPost:
		apiCreateTranslation(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminTranslationHandler dispatches translation operations for admin content data.
func AdminTranslationHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodPost:
		apiCreateAdminTranslation(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiCreateTranslation creates locale-specific content field rows for a content data node.
// It copies the default locale's field values as starting content for each translatable field.
func apiCreateTranslation(w http.ResponseWriter, r *http.Request, c config.Config) error {
	if !c.I18nEnabled() {
		http.Error(w, "i18n is not enabled", http.StatusBadRequest)
		return fmt.Errorf("i18n is not enabled")
	}

	d := db.ConfigDB(c)

	// Parse content data ID from path.
	rawID := r.PathValue("id")
	contentDataID := types.ContentID(rawID)
	if err := contentDataID.Validate(); err != nil {
		utility.DefaultLogger.Error("invalid content data ID", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	// Parse request body.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req translationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.DefaultLogger.Error("failed to decode translation request", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	if req.Locale == "" {
		http.Error(w, "locale is required", http.StatusBadRequest)
		return fmt.Errorf("locale is required")
	}

	// Verify the locale exists and is enabled.
	loc, err := d.GetLocaleByCode(req.Locale)
	if err != nil {
		utility.DefaultLogger.Error("failed to look up locale", err)
		http.Error(w, fmt.Sprintf("locale %q not found", req.Locale), http.StatusBadRequest)
		return err
	}
	if !loc.IsEnabled {
		http.Error(w, fmt.Sprintf("locale %q is not enabled", req.Locale), http.StatusBadRequest)
		return fmt.Errorf("locale %q is not enabled", req.Locale)
	}

	// Get the content data node to find its datatype.
	cd, err := d.GetContentData(contentDataID)
	if err != nil {
		utility.DefaultLogger.Error("failed to get content data", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Get schema fields for the datatype.
	fields, err := d.ListFieldsByDatatypeID(cd.DatatypeID)
	if err != nil {
		utility.DefaultLogger.Error("failed to list fields by datatype", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Get existing fields for the target locale to avoid duplicates.
	nullableContentDataID := types.NullableContentID{ID: contentDataID, Valid: true}
	existingFields, err := d.ListContentFieldsByContentDataAndLocale(nullableContentDataID, req.Locale)
	if err != nil {
		utility.DefaultLogger.Error("failed to list existing locale fields", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Build a set of field IDs that already have rows for the target locale.
	existingFieldSet := make(map[types.FieldID]bool, len(*existingFields))
	for _, ef := range *existingFields {
		if ef.FieldID.Valid {
			existingFieldSet[ef.FieldID.ID] = true
		}
	}

	// Get default locale field values to copy as starting content.
	defaultLocale := c.I18nDefaultLocale()
	defaultFields, err := d.ListContentFieldsByContentDataAndLocale(nullableContentDataID, defaultLocale)
	if err != nil {
		utility.DefaultLogger.Error("failed to list default locale fields", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Build a map from field ID to default locale field value.
	defaultValueMap := make(map[types.FieldID]string, len(*defaultFields))
	for _, df := range *defaultFields {
		if df.FieldID.Valid {
			defaultValueMap[df.FieldID.ID] = df.FieldValue
		}
	}

	// Get the authenticated user for AuthorID.
	user := middleware.AuthenticatedUser(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return fmt.Errorf("authentication required")
	}

	ac := middleware.AuditContextFromRequest(r, c)
	ctx := r.Context()
	now := types.TimestampNow()
	created := 0

	for _, f := range *fields {
		if f.Translatable == 0 {
			continue
		}
		if existingFieldSet[f.FieldID] {
			continue
		}

		fieldValue := defaultValueMap[f.FieldID]

		_, cfErr := d.CreateContentField(ctx, ac, db.CreateContentFieldParams{
			RouteID:       cd.RouteID,
			ContentDataID: nullableContentDataID,
			FieldID:       types.NullableFieldID{ID: f.FieldID, Valid: true},
			FieldValue:    fieldValue,
			Locale:        req.Locale,
			AuthorID:      user.UserID,
			DateCreated:   now,
			DateModified:  now,
		})
		if cfErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("translation: failed to create content field for field %s locale %s", f.FieldID, req.Locale), cfErr)
			http.Error(w, cfErr.Error(), http.StatusInternalServerError)
			return cfErr
		}
		created++
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(translationResponse{
		Locale:        req.Locale,
		FieldsCreated: created,
	})
	return nil
}

// apiCreateAdminTranslation creates locale-specific admin content field rows for an admin content data node.
// It copies the default locale's field values as starting content for each translatable field.
func apiCreateAdminTranslation(w http.ResponseWriter, r *http.Request, c config.Config) error {
	if !c.I18nEnabled() {
		http.Error(w, "i18n is not enabled", http.StatusBadRequest)
		return fmt.Errorf("i18n is not enabled")
	}

	d := db.ConfigDB(c)

	// Parse admin content data ID from path.
	rawID := r.PathValue("id")
	adminContentDataID := types.AdminContentID(rawID)
	if err := adminContentDataID.Validate(); err != nil {
		utility.DefaultLogger.Error("invalid admin content data ID", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	// Parse request body.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req translationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.DefaultLogger.Error("failed to decode translation request", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	if req.Locale == "" {
		http.Error(w, "locale is required", http.StatusBadRequest)
		return fmt.Errorf("locale is required")
	}

	// Verify the locale exists and is enabled.
	loc, err := d.GetLocaleByCode(req.Locale)
	if err != nil {
		utility.DefaultLogger.Error("failed to look up locale", err)
		http.Error(w, fmt.Sprintf("locale %q not found", req.Locale), http.StatusBadRequest)
		return err
	}
	if !loc.IsEnabled {
		http.Error(w, fmt.Sprintf("locale %q is not enabled", req.Locale), http.StatusBadRequest)
		return fmt.Errorf("locale %q is not enabled", req.Locale)
	}

	// Get the admin content data node to find its datatype.
	cd, err := d.GetAdminContentData(adminContentDataID)
	if err != nil {
		utility.DefaultLogger.Error("failed to get admin content data", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Get schema fields for the admin datatype.
	fields, err := d.ListAdminFieldsByDatatypeID(cd.AdminDatatypeID)
	if err != nil {
		utility.DefaultLogger.Error("failed to list admin fields by datatype", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Get existing fields for the target locale to avoid duplicates.
	nullableAdminContentDataID := types.NullableAdminContentID{ID: adminContentDataID, Valid: true}
	existingFields, err := d.ListAdminContentFieldsByContentDataAndLocale(nullableAdminContentDataID, req.Locale)
	if err != nil {
		utility.DefaultLogger.Error("failed to list existing admin locale fields", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Build a set of admin field IDs that already have rows for the target locale.
	existingFieldSet := make(map[types.AdminFieldID]bool, len(*existingFields))
	for _, ef := range *existingFields {
		if ef.AdminFieldID.Valid {
			existingFieldSet[ef.AdminFieldID.ID] = true
		}
	}

	// Get default locale field values to copy as starting content.
	defaultLocale := c.I18nDefaultLocale()
	defaultFields, err := d.ListAdminContentFieldsByContentDataAndLocale(nullableAdminContentDataID, defaultLocale)
	if err != nil {
		utility.DefaultLogger.Error("failed to list default admin locale fields", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Build a map from admin field ID to default locale field value.
	defaultValueMap := make(map[types.AdminFieldID]string, len(*defaultFields))
	for _, df := range *defaultFields {
		if df.AdminFieldID.Valid {
			defaultValueMap[df.AdminFieldID.ID] = df.AdminFieldValue
		}
	}

	// Get the authenticated user for AuthorID.
	user := middleware.AuthenticatedUser(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return fmt.Errorf("authentication required")
	}

	ac := middleware.AuditContextFromRequest(r, c)
	ctx := r.Context()
	now := types.TimestampNow()
	created := 0

	for _, f := range *fields {
		if f.Translatable == 0 {
			continue
		}
		if existingFieldSet[f.AdminFieldID] {
			continue
		}

		fieldValue := defaultValueMap[f.AdminFieldID]

		_, cfErr := d.CreateAdminContentField(ctx, ac, db.CreateAdminContentFieldParams{
			AdminRouteID:       cd.AdminRouteID,
			AdminContentDataID: nullableAdminContentDataID,
			AdminFieldID:       types.NullableAdminFieldID{ID: f.AdminFieldID, Valid: true},
			AdminFieldValue:    fieldValue,
			Locale:             req.Locale,
			AuthorID:           user.UserID,
			DateCreated:        now,
			DateModified:       now,
		})
		if cfErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("translation: failed to create admin content field for field %s locale %s", f.AdminFieldID, req.Locale), cfErr)
			http.Error(w, cfErr.Error(), http.StatusInternalServerError)
			return cfErr
		}
		created++
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(translationResponse{
		Locale:        req.Locale,
		FieldsCreated: created,
	})
	return nil
}
