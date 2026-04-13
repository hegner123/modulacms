package mcp

import "fmt"

// errNoConnection is returned when a tool is invoked but no project is connected.
var errNoConnection = fmt.Errorf("no active connection — use switch_project to connect to a CMS instance")

// proxyBackends wraps a ConnectionManager and lazily delegates to SDKBackends
// built from the current client. When the project is switched, subsequent
// calls automatically use the new client.
type proxyBackends struct {
	cm *ConnectionManager
}

func (p *proxyBackends) backends() (*Backends, error) {
	client := p.cm.Client()
	if client == nil {
		return nil, errNoConnection
	}
	return NewSDKBackends(client), nil
}

// NewProxyBackends returns a Backends struct where every method delegates
// through the ConnectionManager's current client. This enables runtime
// project switching without rebuilding the MCP server.
func NewProxyBackends(cm *ConnectionManager) *Backends {
	pb := &proxyBackends{cm: cm}
	return &Backends{
		Content:           &proxyContentBackend{pb},
		AdminContent:      &proxyAdminContentBackend{pb},
		Schema:            &proxySchemaBackend{pb},
		AdminSchema:       &proxyAdminSchemaBackend{pb},
		Media:             &proxyMediaBackend{pb},
		MediaFolders:      &proxyMediaFolderBackend{pb},
		AdminMedia:        &proxyAdminMediaBackend{pb},
		AdminMediaFolders: &proxyAdminMediaFolderBackend{pb},
		Routes:            &proxyRouteBackend{pb},
		AdminRoutes:       &proxyAdminRouteBackend{pb},
		Users:             &proxyUserBackend{pb},
		RBAC:              &proxyRBACBackend{pb},
		Sessions:          &proxySessionBackend{pb},
		Tokens:            &proxyTokenBackend{pb},
		SSHKeys:           &proxySSHKeyBackend{pb},
		OAuth:             &proxyOAuthBackend{pb},
		Tables:            &proxyTableBackend{pb},
		Plugins:           &proxyPluginBackend{pb},
		Config:            &proxyConfigBackend{pb},
		Import:            &proxyImportBackend{pb},
		Deploy:            &proxyDeployBackend{pb},
		Health:            &proxyHealthBackend{pb},
		Publishing:        &proxyPublishingBackend{pb},
		Versions:          &proxyVersionBackend{pb},
		Webhooks:          &proxyWebhookBackend{pb},
		Locales:           &proxyLocaleBackend{pb},
		Validations:       &proxyValidationBackend{pb},
		Search:            &proxySearchBackend{pb},
		Activity:          &proxyActivityBackend{pb},
		Auth:              &proxyAuthBackend{pb},
	}
}
