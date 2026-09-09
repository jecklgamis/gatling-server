package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

// ForceDownloadHeaders wraps next so file responses are served as an opaque,
// non-executable download rather than letting the browser sniff and render
// the content type. Without this, an uploaded .html/.js file served back
// from /uploads/ would execute in the browser of anyone who opens the link -
// a stored XSS vector, since /upload accepts any file type and serves it
// back publicly.
//
// Requests for a directory (path ending in "/") are left alone: that's
// http.FileServer's own generated listing HTML, not user-uploaded content,
// and forcing it to download too breaks the whole point of these directories
// being browsable.
func ForceDownloadHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", sanitizeFilename(r.URL.Path)))
		next.ServeHTTP(w, r)
	})
}

// sanitizeFilename derives a safe Content-Disposition filename from a
// request path, stripping characters that could break out of the quoted
// header value.
func sanitizeFilename(urlPath string) string {
	name := filepath.Base(urlPath)
	replacer := strings.NewReplacer("\"", "", "\r", "", "\n", "")
	name = replacer.Replace(name)
	if name == "" || name == "." || name == "/" {
		return "download"
	}
	return name
}
