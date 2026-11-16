package httpapi

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// withSPA serves the built frontend from dir for every path outside /api, so
// the browser loads the app and calls the API on one origin (the session
// cookie then needs no cross-site configuration).
//
//   - Existing files are served as-is; /assets/* (content-hashed by Vite) are
//     cached for a year, everything else must be revalidated.
//   - Unknown paths without a file extension get index.html so client-side
//     routes such as /tracker work on reload.
//   - /api/* always goes to the API handler.
func withSPA(dir string, api http.Handler) http.Handler {
	files := http.FileServer(http.Dir(dir))
	index := filepath.Join(dir, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/") {
			api.ServeHTTP(w, r)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}

		clean := path.Clean("/" + r.URL.Path)
		if clean != "/" {
			if info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(clean))); err == nil && !info.IsDir() {
				if strings.HasPrefix(clean, "/assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				} else {
					w.Header().Set("Cache-Control", "no-cache")
				}
				files.ServeHTTP(w, r)
				return
			}
			if path.Ext(clean) != "" {
				http.NotFound(w, r)
				return
			}
		}
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, index)
	})
}
