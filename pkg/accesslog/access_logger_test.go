package accesslog

import (
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNextHandlerInvoked(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	test.Assert(t, err == nil, "unable to create request")

	nextHandlerInvoked := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerInvoked = true
	})
	rr := httptest.NewRecorder()
	handler := AccessLoggerMiddleware(nextHandler)
	handler.ServeHTTP(rr, req)
	test.Assert(t, nextHandlerInvoked, "expecting next handler invocation")
}

func TestNewAccessLoggerMiddlewareLogsThroughGivenLogger(t *testing.T) {
	req, err := http.NewRequest("GET", "/some-path", nil)
	test.Assert(t, err == nil, "unable to create request")
	req.RequestURI = "/some-path"

	var buf strings.Builder
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	rr := httptest.NewRecorder()
	NewAccessLoggerMiddleware(logger)(nextHandler).ServeHTTP(rr, req)

	logged := buf.String()
	test.Assertf(t, strings.Contains(logged, `"msg":"access"`), "expecting access log entry, got %q", logged)
	test.Assertf(t, strings.Contains(logged, `"uri_path":"/some-path"`), "expecting uri_path in log entry, got %q", logged)
}
