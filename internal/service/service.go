// Package service provides a shared business-logic layer for ModulaCMS.
// It sits between transport handlers (HTTP, SSH/TUI, MCP) and the DbDriver
// database interface, eliminating duplicated validation, orchestration,
// and transformation logic across consumers.
//
// The Registry struct is the injection root: constructed at startup, it holds
// all infrastructure dependencies and (eventually) all domain service instances.
// Handlers that have not yet been migrated can access raw dependencies through
// the Registry's getter methods.
package service

import (
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/email"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/publishing"
)

// Registry holds infrastructure dependencies and domain service instances.
// Constructed once at startup and threaded into the router and admin handlers.
type Registry struct {
	driver     db.DbDriver
	mgr        *config.Manager
	pc         *middleware.PermissionCache
	emailSvc   *email.Service
	dispatcher publishing.WebhookDispatcher

	// Domain services — populated in NewRegistry.
	Schema       *SchemaService
	Content      *ContentService
	AdminContent *AdminContentService
	Media        *MediaService
	Routes       *RouteService
	Users        *UserService
	RBAC         *RBACService

	// Phase 5 — constructed post-NewRegistry and assigned externally.
	Plugins  *PluginService
	Webhooks *WebhookService
	Locales  *LocaleService

	// Phase 6 — thin CRUD services.
	Sessions  *SessionService
	Tokens    *TokenService
	SSHKeys   *SSHKeyService
	OAuth     *OAuthService
	Tables    *TableService
	ConfigSvc *ConfigService
	Import    *ImportService
	Deploy    *DeployService
	AuditLog  *AuditLogService
	Backup    *BackupService

	// Phase 7 — auth service.
	Auth   *AuthService
	Search *SearchService
}

// NewRegistry creates a Registry with the given infrastructure dependencies.
func NewRegistry(
	driver db.DbDriver,
	mgr *config.Manager,
	pc *middleware.PermissionCache,
	emailSvc *email.Service,
	dispatcher publishing.WebhookDispatcher,
) *Registry {
	reg := &Registry{
		driver:     driver,
		mgr:        mgr,
		pc:         pc,
		emailSvc:   emailSvc,
		dispatcher: dispatcher,
	}
	reg.Schema = NewSchemaService(driver, driver)
	reg.Content = NewContentService(driver, mgr, dispatcher)
	reg.AdminContent = NewAdminContentService(driver, mgr, dispatcher)
	reg.Media = NewMediaService(driver, mgr)
	reg.Routes = NewRouteService(driver, mgr)
	reg.Users = NewUserService(driver, mgr)
	reg.RBAC = NewRBACService(driver, mgr, pc)
	reg.Sessions = NewSessionService(driver)
	reg.Tokens = NewTokenService(driver)
	reg.SSHKeys = NewSSHKeyService(driver)
	reg.OAuth = NewOAuthService(driver)
	reg.Tables = NewTableService(driver)
	reg.ConfigSvc = NewConfigService(mgr)
	reg.Import = NewImportService(driver, mgr)
	reg.Deploy = NewDeployService(driver, mgr)
	reg.AuditLog = NewAuditLogService(driver)
	reg.Backup = NewBackupService(mgr, driver)
	reg.Auth = NewAuthService(driver, mgr, emailSvc)
	return reg
}

// Driver returns the database driver for unmigrated handlers.
func (r *Registry) Driver() db.DbDriver {
	return r.driver
}

// Config returns the current configuration snapshot.
func (r *Registry) Config() (*config.Config, error) {
	return r.mgr.Config()
}

// Manager returns the config manager for unmigrated handlers.
func (r *Registry) Manager() *config.Manager {
	return r.mgr
}

// PermissionCache returns the RBAC permission cache.
func (r *Registry) PermissionCache() *middleware.PermissionCache {
	return r.pc
}

// EmailService returns the email service for unmigrated handlers.
func (r *Registry) EmailService() *email.Service {
	return r.emailSvc
}

// Dispatcher returns the webhook dispatcher for unmigrated handlers.
func (r *Registry) Dispatcher() publishing.WebhookDispatcher {
	return r.dispatcher
}
