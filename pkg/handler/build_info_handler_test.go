package handler

import (
	"encoding/json"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildInfoHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/buildInfo", nil)
	test.Assert(t, err == nil, "unable create request")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(BuildInfoHandler)
	handler.ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusOK, "unexpected status code %v", rr.Code)
	test.Assertf(t, rr.Header().Get("Content-Type") == "application/json",
		"unexpected content type %s", rr.Header().Get("Content-Type"))

	var entity map[string]string
	_ = json.Unmarshal(rr.Body.Bytes(), &entity)
	test.Assert(t, entity["name"] == "gatling-server", "unexpected name")
	test.Assert(t, entity["version"] == "", "unexpected version")
	test.Assert(t, entity["branch"] == "", "unexpected branch")
}
