package accesslog

import (
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"net/http"
	"net/http/httptest"
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
