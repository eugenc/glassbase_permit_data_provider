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
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		up := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		try := up
		if try == "" || try == "." {
			try = "index.html"
		} else if _, err := fs.Stat(sub, try); err != nil {
			try = "index.html"
		}

		req := r.Clone(r.Context())
		u := *r.URL
		u.Path = "/" + try
		req.URL = &u

		fileServer.ServeHTTP(w, req)
	})
}
