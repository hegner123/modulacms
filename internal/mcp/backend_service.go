package mcp

import (
	"github.com/hegner123/modulacms/internal/service"
)

// NewServiceBackends creates a Backends wired to a service.Registry for
// in-process (direct mode) operation. Every domain adapter wraps the
// corresponding service with JSON marshaling. Audit context is built
// per-call from the authenticated user in context via AuditContextFromMCP.
func NewServiceBackends(svc *service.Registry) *Backends {
	return &Backends{
		Content:           &svcContentBackend{svc: svc},
		AdminContent:      &svcAdminContentBackend{svc: svc},
		Schema:            &svcSchemaBackend{svc: svc},
		AdminSchema:       &svcAdminSchemaBackend{svc: svc},
		Media:             &svcMediaBackend{svc: svc},
		MediaFolders:      &svcMediaFolderBackend{svc: svc},
		AdminMedia:        &svcAdminMediaBackend{svc: svc},
		AdminMediaFolders: &svcAdminMediaFolderBackend{svc: svc},
		Routes:            &svcRouteBackend{svc: svc},
		AdminRoutes:       &svcAdminRouteBackend{svc: svc},
		Users:             &svcUserBackend{svc: svc},
		RBAC:              &svcRBACBackend{svc: svc},
		Sessions:          &svcSessionBackend{svc: svc},
		Tokens:            &svcTokenBackend{svc: svc},
		SSHKeys:           &svcSSHKeyBackend{svc: svc},
		OAuth:             &svcOAuthBackend{svc: svc},
		Tables:            &svcTableBackend{svc: svc},
		Plugins:           &svcPluginBackend{svc: svc},
		Config:            &svcConfigBackend{svc: svc},
		Import:            &svcImportBackend{svc: svc},
		Deploy:            &svcDeployBackend{svc: svc},
		Health:            &svcHealthBackend{svc: svc},
		Publishing:        &svcPublishingBackend{svc: svc},
		Versions:          &svcVersionBackend{svc: svc},
		Webhooks:          &svcWebhookBackend{svc: svc},
		Locales:           &svcLocaleBackend{svc: svc},
		Validations:       &svcValidationBackend{svc: svc},
		Search:            &svcSearchBackend{svc: svc},
		Activity:          &svcActivityBackend{svc: svc},
		Auth:              &svcAuthBackend{svc: svc},
	}
}
