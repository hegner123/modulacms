//go:build !dev

package admin

import "embed"

//go:embed static/*
var staticFiles embed.FS
