package handlers

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/service"
)

// stageColor maps the environment stage to an accent color for the favicon.
// The color reflects the stage (local/dev/staging/prod), not the deployment
// variant (native vs docker).
func stageColor(stage string) string {
	switch stage {
	case "production":
		return "#ef4444" // red
	case "staging":
		return "#f59e0b" // amber
	case "development":
		return "#22c55e" // green
	case "local":
		return "#3b82f6" // blue
	default:
		return "#3b82f6" // blue
	}
}

// FaviconHandler serves a dynamic SVG favicon colored by the config environment.
func FaviconHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		color := "#3b82f6"
		cfg, err := svc.Config()
		if err == nil {
			color = stageColor(cfg.Environment.Stage())
		}

		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">
  <defs>
    <clipPath id="bg-clip">
      <rect width="64" height="64" rx="12"/>
    </clipPath>
  </defs>
  <rect width="64" height="64" rx="12" fill="#000"/>
  <g clip-path="url(#bg-clip)">
    <rect x="38" y="0" width="26" height="26" rx="3" fill="` + color + `"/>
    <rect x="0" y="26" width="38" height="38" rx="3" fill="` + color + `"/>
  </g>
</svg>`))
	}
}
