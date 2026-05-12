package static

import "embed"

// FS holds production frontend assets (built by Vite into this tree).
//
//go:embed all:web
var FS embed.FS
