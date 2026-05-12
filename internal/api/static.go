package api

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/echayko/glassbase_permit_data_provider/internal/static"
)

// StaticHandler serves the embedded React SPA (production build).
func StaticHandler() http.Handler {
	sub, err := fs.Sub(static.FS, "web")
	if err != nil {
		panic("static FS sub web: " + err.Error())
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		up := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		name := up
		if name == "" || name == "." {
			name = "index.html"
		} else if _, err := fs.Stat(sub, name); err != nil {
			name = "index.html"
		}
		http.ServeFileFS(w, r, sub, name)
	})
}
