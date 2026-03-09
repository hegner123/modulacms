package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// RouteService manages route CRUD with slug uniqueness, cascade-check on delete,
// and consistent validation for both public and admin routes.
type RouteService struct {
	driver db.DbDriver
	mgr    *config.Manager
}

// NewRouteService creates a RouteService with the given dependencies.
func NewRouteService(driver db.DbDriver, mgr *config.Manager) *RouteService {
	return &RouteService{driver: driver, mgr: mgr}
}

// CreateRouteInput holds the caller-provided fields for creating a route.
type CreateRouteInput struct {
	Slug     string               `json:"slug"`
	Title    string               `json:"title"`
	Status   int64                `json:"status"`
	AuthorID types.NullableUserID `json:"author_id"`
}

// UpdateRouteInput holds the caller-provided fields for updating a route.
// The caller provides RouteID; the service fetches the existing record
// to populate Slug_2 and preserve DateCreated/AuthorID.
type UpdateRouteInput struct {
	RouteID  types.RouteID        `json:"route_id"`
	Slug     string               `json:"slug"`
	Title    string               `json:"title"`
	Status   int64                `json:"status"`
	AuthorID types.NullableUserID `json:"author_id"`
}

// --- Public Route Methods ---

// CreateRoute validates input, checks slug uniqueness, and creates a new route.
func (s *RouteService) CreateRoute(ctx context.Context, ac audited.AuditContext, input CreateRouteInput) (*db.Routes, error) {
	if err := validateRouteInput(input.Slug, input.Title); err != nil {
		return nil, err
	}

	if err := s.checkSlugUniqueness(input.Slug, ""); err != nil {
		return nil, err
	}

	now := types.NewTimestamp(time.Now().UTC())
	created, err := s.driver.CreateRoute(ctx, ac, db.CreateRouteParams{
		Slug:         types.Slug(input.Slug),
		Title:        input.Title,
		Status:       input.Status,
		AuthorID:     input.AuthorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		return nil, fmt.Errorf("create route: %w", err)
	}
	return created, nil
}

// UpdateRoute validates input, checks slug uniqueness (excluding self),
// fetches the existing record to populate Slug_2, and updates the route.
func (s *RouteService) UpdateRoute(ctx context.Context, ac audited.AuditContext, input UpdateRouteInput) (*db.Routes, error) {
	if err := validateRouteInput(input.Slug, input.Title); err != nil {
		return nil, err
	}

	existing, err := s.driver.GetRoute(input.RouteID)
	if err != nil {
		return nil, &NotFoundError{Resource: "route", ID: string(input.RouteID)}
	}

	if err := s.checkSlugUniqueness(input.Slug, string(existing.Slug)); err != nil {
		return nil, err
	}

	authorID := input.AuthorID
	if !authorID.Valid {
		authorID = existing.AuthorID
	}

	_, err = s.driver.UpdateRoute(ctx, ac, db.UpdateRouteParams{
		Slug:         types.Slug(input.Slug),
		Title:        input.Title,
		Status:       input.Status,
		AuthorID:     authorID,
		DateCreated:  existing.DateCreated,
		DateModified: types.NewTimestamp(time.Now().UTC()),
		Slug_2:       existing.Slug,
	})
	if err != nil {
		return nil, fmt.Errorf("update route: %w", err)
	}

	// Re-fetch to return fresh data. Use GetRouteID to find by new slug,
	// then GetRoute by ID.
	newID, err := s.driver.GetRouteID(input.Slug)
	if err != nil {
		return nil, fmt.Errorf("fetch updated route ID: %w", err)
	}
	updated, err := s.driver.GetRoute(*newID)
	if err != nil {
		return nil, fmt.Errorf("fetch updated route: %w", err)
	}
	return updated, nil
}

// DeleteRoute checks that no content references this route, then deletes it.
func (s *RouteService) DeleteRoute(ctx context.Context, ac audited.AuditContext, id types.RouteID) error {
	content, err := s.driver.ListContentDataByRoute(types.NullableRouteID{ID: id, Valid: true})
	if err != nil {
		return fmt.Errorf("check route content: %w", err)
	}
	if content != nil && len(*content) > 0 {
		return &ConflictError{
			Resource: "route",
			ID:       string(id),
			Detail:   fmt.Sprintf("route has %d content item(s) attached", len(*content)),
		}
	}

	if err := s.driver.DeleteRoute(ctx, ac, id); err != nil {
		return fmt.Errorf("delete route: %w", err)
	}
	return nil
}

// GetRoute retrieves a route by ID with NotFoundError mapping.
func (s *RouteService) GetRoute(ctx context.Context, id types.RouteID) (*db.Routes, error) {
	route, err := s.driver.GetRoute(id)
	if err != nil {
		return nil, &NotFoundError{Resource: "route", ID: string(id)}
	}
	return route, nil
}

// GetRouteFull retrieves a composed route view with author and content tree.
func (s *RouteService) GetRouteFull(ctx context.Context, id types.RouteID) (*db.RouteFullView, error) {
	view, err := db.AssembleRouteFullView(s.driver, id)
	if err != nil {
		return nil, &NotFoundError{Resource: "route", ID: string(id)}
	}
	return view, nil
}

// ListRoutes returns all routes.
func (s *RouteService) ListRoutes(ctx context.Context) (*[]db.Routes, error) {
	return s.driver.ListRoutes()
}

// ListRoutesPaginated returns routes with pagination.
func (s *RouteService) ListRoutesPaginated(ctx context.Context, limit, offset int64) (*[]db.Routes, *int64, error) {
	items, err := s.driver.ListRoutesPaginated(db.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, nil, fmt.Errorf("list routes paginated: %w", err)
	}
	total, err := s.driver.CountRoutes()
	if err != nil {
		return nil, nil, fmt.Errorf("count routes: %w", err)
	}
	return items, total, nil
}

// --- Admin Route Methods ---

// CreateAdminRoute validates input, checks slug uniqueness, and creates a new admin route.
func (s *RouteService) CreateAdminRoute(ctx context.Context, ac audited.AuditContext, input CreateRouteInput) (*db.AdminRoutes, error) {
	if err := validateRouteInput(input.Slug, input.Title); err != nil {
		return nil, err
	}

	if err := s.checkAdminSlugUniqueness(input.Slug, ""); err != nil {
		return nil, err
	}

	now := types.NewTimestamp(time.Now().UTC())
	created, err := s.driver.CreateAdminRoute(ctx, ac, db.CreateAdminRouteParams{
		Slug:         types.Slug(input.Slug),
		Title:        input.Title,
		Status:       input.Status,
		AuthorID:     input.AuthorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		return nil, fmt.Errorf("create admin route: %w", err)
	}
	return created, nil
}

// UpdateAdminRoute validates input, checks slug uniqueness (excluding self),
// fetches the existing record, and updates the admin route.
func (s *RouteService) UpdateAdminRoute(ctx context.Context, ac audited.AuditContext, input UpdateRouteInput) (*db.AdminRoutes, error) {
	if err := validateRouteInput(input.Slug, input.Title); err != nil {
		return nil, err
	}

	// Admin routes are fetched by ID, but GetAdminRoute uses slug.
	// We need to look up the existing route. Since UpdateRouteInput has RouteID
	// (which is actually AdminRouteID cast), we use GetAdminRouteByID if available,
	// or iterate routes. For now, the caller provides the existing slug via the
	// input.Slug field if unchanged. We fetch all admin routes to find by ID.
	existing, err := s.findAdminRouteByID(types.AdminRouteID(input.RouteID))
	if err != nil {
		return nil, err
	}

	if err := s.checkAdminSlugUniqueness(input.Slug, string(existing.Slug)); err != nil {
		return nil, err
	}

	authorID := input.AuthorID
	if !authorID.Valid {
		authorID = existing.AuthorID
	}

	_, err = s.driver.UpdateAdminRoute(ctx, ac, db.UpdateAdminRouteParams{
		Slug:         types.Slug(input.Slug),
		Title:        input.Title,
		Status:       input.Status,
		AuthorID:     authorID,
		DateCreated:  existing.DateCreated,
		DateModified: types.NewTimestamp(time.Now().UTC()),
		Slug_2:       existing.Slug,
	})
	if err != nil {
		return nil, fmt.Errorf("update admin route: %w", err)
	}

	// Re-fetch to return fresh data (admin routes are fetched by slug).
	updated, err := s.driver.GetAdminRoute(types.Slug(input.Slug))
	if err != nil {
		return nil, fmt.Errorf("fetch updated admin route: %w", err)
	}
	return updated, nil
}

// DeleteAdminRoute checks that no admin content references this route, then deletes it.
func (s *RouteService) DeleteAdminRoute(ctx context.Context, ac audited.AuditContext, id types.AdminRouteID) error {
	content, err := s.driver.ListAdminContentDataByRoute(types.NullableAdminRouteID{ID: id, Valid: true})
	if err != nil {
		return fmt.Errorf("check admin route content: %w", err)
	}
	if content != nil && len(*content) > 0 {
		return &ConflictError{
			Resource: "admin_route",
			ID:       string(id),
			Detail:   fmt.Sprintf("route has %d content item(s) attached", len(*content)),
		}
	}

	if err := s.driver.DeleteAdminRoute(ctx, ac, id); err != nil {
		return fmt.Errorf("delete admin route: %w", err)
	}
	return nil
}

// GetAdminRoute retrieves an admin route by ID with NotFoundError mapping.
func (s *RouteService) GetAdminRoute(ctx context.Context, id types.AdminRouteID) (*db.AdminRoutes, error) {
	route, err := s.findAdminRouteByID(id)
	if err != nil {
		return nil, err
	}
	return route, nil
}

// GetAdminRouteBySlug retrieves an admin route by slug with NotFoundError mapping.
func (s *RouteService) GetAdminRouteBySlug(ctx context.Context, slug types.Slug) (*db.AdminRoutes, error) {
	route, err := s.driver.GetAdminRoute(slug)
	if err != nil {
		return nil, &NotFoundError{Resource: "admin_route", ID: string(slug)}
	}
	return route, nil
}

// ListAdminRoutes returns all admin routes.
func (s *RouteService) ListAdminRoutes(ctx context.Context) (*[]db.AdminRoutes, error) {
	return s.driver.ListAdminRoutes()
}

// ListAdminRoutesPaginated returns admin routes with pagination.
func (s *RouteService) ListAdminRoutesPaginated(ctx context.Context, limit, offset int64) (*[]db.AdminRoutes, *int64, error) {
	items, err := s.driver.ListAdminRoutesPaginated(db.PaginationParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, nil, fmt.Errorf("list admin routes paginated: %w", err)
	}
	total, err := s.driver.CountAdminRoutes()
	if err != nil {
		return nil, nil, fmt.Errorf("count admin routes: %w", err)
	}
	return items, total, nil
}

// ListOrderedAdminRoutes returns admin routes sorted by the "Order" field
// on each route's root admin content node. Routes without an Order value appear last.
func (s *RouteService) ListOrderedAdminRoutes(ctx context.Context) ([]db.AdminRoutes, error) {
	routes, err := s.driver.ListAdminRoutes()
	if err != nil {
		return nil, fmt.Errorf("list admin routes: %w", err)
	}

	allFields, err := s.driver.ListAdminFields()
	if err != nil {
		return nil, fmt.Errorf("list admin fields: %w", err)
	}

	var orderFieldID types.AdminFieldID
	for _, f := range *allFields {
		if f.Label == "Order" {
			orderFieldID = f.AdminFieldID
			break
		}
	}

	// If no Order field exists, return routes unsorted.
	if orderFieldID.IsZero() {
		return *routes, nil
	}

	type orderedRoute struct {
		Route db.AdminRoutes
		Order int
	}
	ordered := make([]orderedRoute, 0, len(*routes))

	for _, route := range *routes {
		nullableRouteID := types.NullableAdminRouteID{ID: route.AdminRouteID, Valid: true}
		order := math.MaxInt32

		contentData, cdErr := s.driver.ListAdminContentDataByRoute(nullableRouteID)
		if cdErr != nil {
			ordered = append(ordered, orderedRoute{Route: route, Order: order})
			continue
		}

		var rootContentDataID string
		for _, cd := range *contentData {
			if !cd.ParentID.Valid {
				rootContentDataID = cd.AdminContentDataID.String()
				break
			}
		}

		if rootContentDataID != "" {
			contentFields, cfErr := s.driver.ListAdminContentFieldsByRoute(nullableRouteID)
			if cfErr != nil {
				ordered = append(ordered, orderedRoute{Route: route, Order: order})
				continue
			}

			for _, cf := range *contentFields {
				if cf.AdminContentDataID.String() == rootContentDataID &&
					cf.AdminFieldID.Valid &&
					cf.AdminFieldID.ID == orderFieldID {
					parsed, parseErr := strconv.Atoi(cf.AdminFieldValue)
					if parseErr == nil {
						order = parsed
					}
					break
				}
			}
		}

		ordered = append(ordered, orderedRoute{Route: route, Order: order})
	}

	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Order < ordered[j].Order
	})

	result := make([]db.AdminRoutes, len(ordered))
	for i, o := range ordered {
		result[i] = o.Route
	}
	return result, nil
}

// --- Private Helpers ---

// validateRouteInput validates slug format and title presence.
func validateRouteInput(slug, title string) *ValidationError {
	ve := &ValidationError{}
	if slug == "" {
		ve.Add("slug", "Slug is required")
	} else if err := types.Slug(slug).Validate(); err != nil {
		ve.Add("slug", err.Error())
	}
	if title == "" {
		ve.Add("title", "Title is required")
	}
	if ve.HasErrors() {
		return ve
	}
	return nil
}

// checkSlugUniqueness checks that no public route with the given slug exists.
// currentSlug is the route's existing slug (empty for create); if the slug
// hasn't changed, no conflict is raised.
func (s *RouteService) checkSlugUniqueness(slug, currentSlug string) error {
	if slug == currentSlug {
		return nil
	}
	existingID, err := s.driver.GetRouteID(slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("check slug uniqueness: %w", err)
	}
	if existingID != nil {
		return &ConflictError{Resource: "route", ID: slug, Detail: "slug already exists"}
	}
	return nil
}

// checkAdminSlugUniqueness checks that no admin route with the given slug exists.
// Uses GetAdminRoute (slug-based lookup). currentSlug is excluded for updates.
func (s *RouteService) checkAdminSlugUniqueness(slug, currentSlug string) error {
	if slug == currentSlug {
		return nil
	}
	existing, err := s.driver.GetAdminRoute(types.Slug(slug))
	if err != nil {
		// Not found = available.
		return nil
	}
	if existing != nil {
		return &ConflictError{Resource: "admin_route", ID: slug, Detail: "slug already exists"}
	}
	return nil
}

// findAdminRouteByID looks up an admin route by its AdminRouteID.
// Since GetAdminRoute takes a slug, we list all admin routes and find by ID.
func (s *RouteService) findAdminRouteByID(id types.AdminRouteID) (*db.AdminRoutes, error) {
	routes, err := s.driver.ListAdminRoutes()
	if err != nil {
		return nil, fmt.Errorf("list admin routes for ID lookup: %w", err)
	}
	for _, r := range *routes {
		if r.AdminRouteID == id {
			return &r, nil
		}
	}
	return nil, &NotFoundError{Resource: "admin_route", ID: string(id)}
}
