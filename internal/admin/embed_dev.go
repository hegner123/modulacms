//go:build dev

package admin

import (
	"io/fs"
	"os"
)

// staticFiles serves from disk in dev mode so CSS/JS changes are instant.
var staticFiles fs.FS = os.DirFS("internal/admin")
