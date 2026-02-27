package handlers

import (
    "net/http"

    "github.com/hegner123/modulacms/internal/admin/pages"
)

func DemoHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        layout := NewAdminData(r, "Demo")
        RenderNav(w, r, "Demo", pages.DemoContent(), pages.Demo(layout))
    }
}
