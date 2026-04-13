package mcp

import (
	modula "github.com/hegner123/modulacms/sdks/go"
)

// NewSDKBackends returns a Backends struct with all 19 fields populated by
// SDK adapter implementations that delegate to the given client over HTTP.
func NewSDKBackends(client *modula.Client) *Backends {
	return &Backends{
		Content:           &sdkContentBackend{client: client},
		AdminContent:      &sdkAdminContentBackend{client: client},
		Schema:            &sdkSchemaBackend{client: client},
		AdminSchema:       &sdkAdminSchemaBackend{client: client},
		Media:             &sdkMediaBackend{client: client},
		MediaFolders:      &sdkMediaFolderBackend{client: client},
		AdminMedia:        &sdkAdminMediaBackend{client: client},
		AdminMediaFolders: &sdkAdminMediaFolderBackend{client: client},
		Routes:            &sdkRouteBackend{client: client},
		AdminRoutes:       &sdkAdminRouteBackend{client: client},
		Users:             &sdkUserBackend{client: client},
		RBAC:              &sdkRBACBackend{client: client},
		Sessions:          &sdkSessionBackend{client: client},
		Tokens:            &sdkTokenBackend{client: client},
		SSHKeys:           &sdkSSHKeyBackend{client: client},
		OAuth:             &sdkOAuthBackend{client: client},
		Tables:            &sdkTableBackend{client: client},
		Plugins:           &sdkPluginBackend{client: client},
		Config:            &sdkConfigBackend{client: client},
		Import:            &sdkImportBackend{client: client},
		Deploy:            &sdkDeployBackend{client: client},
		Health:            &sdkHealthBackend{client: client},
		Publishing:        &sdkPublishingBackend{client: client},
		Versions:          &sdkVersionBackend{client: client},
		Webhooks:          &sdkWebhookBackend{client: client},
		Locales:           &sdkLocaleBackend{client: client},
		Validations:       &sdkValidationBackend{client: client},
		Search:            &sdkSearchBackend{client: client},
		Activity:          &sdkActivityBackend{client: client},
		Auth:              &sdkAuthBackend{client: client},
	}
}
