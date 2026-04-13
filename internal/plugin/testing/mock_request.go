package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/hegner123/modulacms/internal/plugin"
)

// MockRule defines a canned response for an outbound HTTP request.
type MockRule struct {
	Method      string         // exact match, case-insensitive
	HostPath    string         // host+path to match (scheme and query ignored)
	Response    map[string]any // must contain "status" (int); optional "json" or "body"
}

// MockRequestEngine implements plugin.OutboundExecutor with configurable rules.
// Rules are matched in registration order; first match wins.
type MockRequestEngine struct {
	mu    sync.Mutex
	rules []MockRule
}

// Execute finds the first matching rule and returns its response.
// Returns an error if no rule matches.
func (m *MockRequestEngine) Execute(ctx context.Context, pluginName, method, urlStr string, opts plugin.OutboundRequestOpts) (map[string]any, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("mock: invalid URL %q: %w", urlStr, err)
	}
	hostPath := parsed.Host + parsed.Path

	for _, rule := range m.rules {
		if !strings.EqualFold(rule.Method, method) {
			continue
		}
		if rule.HostPath != hostPath {
			continue
		}
		return m.buildResponse(rule.Response), nil
	}

	return nil, fmt.Errorf("mock: no rule matched %s %s", method, urlStr)
}

// AddRule registers a mock rule. The urlPattern is matched against host+path
// of the request URL (scheme and query parameters are ignored).
func (m *MockRequestEngine) AddRule(method, urlPattern string, response map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rules = append(m.rules, MockRule{
		Method:   method,
		HostPath: urlPattern,
		Response: response,
	})
}

// ClearRules removes all registered mock rules.
func (m *MockRequestEngine) ClearRules() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rules = nil
}

// MockOutboundOpts returns a zero-value OutboundRequestOpts for use in tests.
func MockOutboundOpts() plugin.OutboundRequestOpts {
	return plugin.OutboundRequestOpts{}
}

func (m *MockRequestEngine) buildResponse(resp map[string]any) map[string]any {
	status := 200
	if s, ok := resp["status"]; ok {
		switch v := s.(type) {
		case int:
			status = v
		case float64:
			status = int(v)
		case int64:
			status = int(v)
		}
	}

	result := map[string]any{
		"status":  status,
		"headers": map[string]string{},
	}

	if jsonVal, ok := resp["json"]; ok {
		result["json"] = jsonVal
		bodyBytes, err := json.Marshal(jsonVal)
		if err == nil {
			result["body"] = string(bodyBytes)
		}
	} else if bodyVal, ok := resp["body"]; ok {
		result["body"] = bodyVal
	} else {
		result["body"] = ""
	}

	return result
}
