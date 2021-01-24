package uploader

import (
	"github.com/google/uuid"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateMultipartRequestForNonExistingFile(t *testing.T) {
	_, err := CreateMultipartRequest("http://some-url", "testdata/non-existing-file.txt",
		someRequestData())
	test.Assert(t, err != nil, "expecting to fail request creation")
}

func TestCreateMultipartRequest(t *testing.T) {
	req, err := CreateMultipartRequest("http://some-url",
		"testdata/SingleFileExampleSimulation.scala", someRequestData())
	test.Assert(t, err == nil, "unable to create request")
	validateMultiPartRequest(t, req)
}

func TestUploadFile(t *testing.T) {
	var requestReceived = false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		requestReceived = true
	})
	server := httptest.NewServer(handler)
	resp, err := UploadFile(server.URL, "testdata/SingleFileExampleSimulation.scala", someRequestData())
	test.Assertf(t, resp.StatusCode == http.StatusOK, "expecting 200")
	test.Assertf(t, err == nil, "failed to upload file")
	test.Assert(t, requestReceived, "server did not receive POST request")
}

func TestUploadFileFailureOn5xx(t *testing.T) {
	var requestReceived = false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		requestReceived = true
	})
	server := httptest.NewServer(handler)
	resp, err := UploadFile(server.URL, "testdata/SingleFileExampleSimulation.scala", someRequestData())
	test.Assertf(t, err == nil, "failed to upload file")
	test.Assertf(t, resp.StatusCode == http.StatusInternalServerError, "expecting 500")
	test.Assert(t, requestReceived, "server did not receive POST request")
}

func TestUploadFileFailureOnUnknownHost(t *testing.T) {
	_, err := UploadFile("http://"+uuid.New().String()[0:8], "testdata/SingleFileExampleSimulation.scala", someRequestData())
	test.Assertf(t, err != nil, "failed to upload file")
}

func TestUploadNonExistentFile(t *testing.T) {
	_, err := UploadFile("http://localhost:8080", "testdata/non-existent-file.txt", someRequestData())
	test.Assertf(t, err != nil, "failed to upload file")

}

func validateMultiPartRequest(t *testing.T, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	test.Assert(t, strings.Contains(contentType, "multipart/form-data"), "unexpected content type")
	test.Assert(t, req.Body != nil, "unable to create request")
}

func someRequestData() map[string]string {
	return map[string]string{
		"simulation": "gatling.test.example.simulation.SingleFileExampleSimulation",
		"javaOpts":   "-DbaseUrl=http://localhost:8080 -DdurationMin=0.10 -DrequestPersecond=1",
	}
}
