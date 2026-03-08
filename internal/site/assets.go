package site

import "embed"

//go:embed static/* templates/*.html
var assets embed.FS
