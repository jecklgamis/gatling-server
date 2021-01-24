package handler

import (
	"encoding/json"
	"fmt"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLivenessProbeHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/probe/ready", nil)
	test.Assert(t, err == nil, "unable create request")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(LivenessProbeHandler)
	handler.ServeHTTP(rr, req)
	test.Assert(t, rr.Code == http.StatusOK, "unexpected status code")
	test.Assert(t, rr.Header().Get("Content-Type") == "application/json",
		fmt.Sprintf("unexpected content type %s", rr.Header().Get("Content-Type")))

	var entity map[string]string
	_ = json.Unmarshal(rr.Body.Bytes(), &entity)
	test.Assert(t, entity["message"] == "I'm alive!", fmt.Sprintf("unexpected message : %s", entity["message"]))
}
