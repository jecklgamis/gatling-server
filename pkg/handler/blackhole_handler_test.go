package handler

import (
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBlackholeHandler(t *testing.T) {
	req, err := http.NewRequest("POST", "/blackhole", nil)
	test.Assert(t, err == nil, "unable to create request")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(BlackholeHandler)
	handler.ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusOK, "unexpected status code %v", rr.Code)
}
