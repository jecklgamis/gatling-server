package handler

import (
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAuthAllowsValidToken(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	req, _ := http.NewRequest("GET", "/workspace/some-task-id/console.log", nil)
	req.Header.Set("Authorization", "Bearer "+someApiToken)
	rr := httptest.NewRecorder()
	RequireAuth(inner, someApiToken).ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusOK, "unexpected status code %d", rr.Code)
}

func TestRequireAuthRejectsMissingToken(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	req, _ := http.NewRequest("GET", "/workspace/some-task-id/console.log", nil)
	rr := httptest.NewRecorder()
	RequireAuth(inner, someApiToken).ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusUnauthorized, "unexpected status code %d", rr.Code)
}

func TestRequireAuthRateLimitsAfterTooManyFailures(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := RequireAuth(inner, someApiToken)
	for i := 0; i < authMaxFailures; i++ {
		req, _ := http.NewRequest("GET", "/workspace/some-task-id/console.log", nil)
		req.Header.Set("Authorization", "Bearer wrong-token")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		test.Assertf(t, rr.Code == http.StatusUnauthorized, "unexpected status code %d on attempt %d", rr.Code, i)
	}
	req, _ := http.NewRequest("GET", "/workspace/some-task-id/console.log", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusTooManyRequests, "unexpected status code %d", rr.Code)
}
