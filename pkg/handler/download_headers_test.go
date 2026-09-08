package handler

import (
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestForceDownloadHeadersOverridesSniffedContentType(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<script>alert(1)</script>"))
	})
	req, _ := http.NewRequest("GET", "/uploads/some-id/evil.html", nil)
	rr := httptest.NewRecorder()
	ForceDownloadHeaders(inner).ServeHTTP(rr, req)
	test.Assertf(t, rr.Header().Get("Content-Type") == "application/octet-stream",
		"unexpected content type %q", rr.Header().Get("Content-Type"))
	test.Assertf(t, rr.Header().Get("X-Content-Type-Options") == "nosniff",
		"expecting nosniff header")
	test.Assertf(t, rr.Header().Get("Content-Disposition") == `attachment; filename="evil.html"`,
		"unexpected content disposition %q", rr.Header().Get("Content-Disposition"))
}

func TestSanitizeFilenameStripsQuotesAndNewlines(t *testing.T) {
	name := sanitizeFilename(`/uploads/some-id/foo"bar` + "\r\n.txt")
	test.Assertf(t, !strings.ContainsAny(name, "\"\r\n"), "unexpected sanitized name %q", name)
}
