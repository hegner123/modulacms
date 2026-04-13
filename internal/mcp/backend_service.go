package mcp

import (
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/service"
)

// NewServiceBackends creates a Backends wired to a service.Registry for
// in-process (direct mode) operation. Every domain adapter wraps the
// corresponding service with JSON marshaling.
func NewServiceBackends(svc *service.Registry, ac audited.AuditContext) *Backends {
	return &Backends{
		Content:           &svcContentBackend{svc: svc, ac: ac},
		AdminContent:      &svcAdminContentBackend{svc: svc, ac: ac},
		Schema:            &svcSchemaBackend{svc: svc, ac: ac},
		AdminSchema:       &svcAdminSchemaBackend{svc: svc, ac: ac},
		Media:             &svcMediaBackend{svc: svc, ac: ac},
		MediaFolders:      &svcMediaFolderBackend{svc: svc, ac: ac},
		AdminMedia:        &svcAdminMediaBackend{svc: svc, ac: ac},
		AdminMediaFolders: &svcAdminMediaFolderBackend{svc: svc, ac: ac},
		Routes:            &svcRouteBackend{svc: svc, ac: ac},
		AdminRoutes:       &svcAdminRouteBackend{svc: svc, ac: ac},
		Users:             &svcUserBackend{svc: svc, ac: ac},
		RBAC:              &svcRBACBackend{svc: svc, ac: ac},
		Sessions:          &svcSessionBackend{svc: svc, ac: ac},
		Tokens:            &svcTokenBackend{svc: svc, ac: ac},
		SSHKeys:           &svcSSHKeyBackend{svc: svc, ac: ac},
		OAuth:             &svcOAuthBackend{svc: svc, ac: ac},
		Tables:            &svcTableBackend{svc: svc, ac: ac},
		Plugins:           &svcPluginBackend{svc: svc},
		Config:            &svcConfigBackend{svc: svc},
		Import:            &svcImportBackend{svc: svc, ac: ac},
		Deploy:            &svcDeployBackend{svc: svc},
		Health:            &svcHealthBackend{svc: svc},
		Publishing:        &svcPublishingBackend{svc: svc, ac: ac},
		Versions:          &svcVersionBackend{svc: svc, ac: ac},
		Webhooks:          &svcWebhookBackend{svc: svc, ac: ac},
		Locales:           &svcLocaleBackend{svc: svc, ac: ac},
		Validations:       &svcValidationBackend{svc: svc, ac: ac},
		Search:            &svcSearchBackend{svc: svc},
		Activity:          &svcActivityBackend{svc: svc},
		Auth:              &svcAuthBackend{svc: svc, ac: ac},
	}
}
