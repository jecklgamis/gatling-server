package handler

import (
	"encoding/json"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"github.com/jecklgamis/gatling-server/pkg/uploader"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileUploadWithInvalidMethod(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost:8080", strings.NewReader("some-body"))
	req.Header.Set("Authorization", "Bearer "+someApiToken)
	createFileUploadHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusMethodNotAllowed, "unexpected status code %d", rr.Code)
}

func TestFileUploadWithMissingApiToken(t *testing.T) {
	req := createFileUploadHttpRequest(t, "testdata/some.txt", "")
	rr := httptest.NewRecorder()
	createFileUploadHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusUnauthorized, "unexpected status code %d", rr.Code)
}

func TestFileUploadWithInvalidApiToken(t *testing.T) {
	req := createFileUploadHttpRequest(t, "testdata/some.txt", "some-invalid-token")
	rr := httptest.NewRecorder()
	createFileUploadHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusUnauthorized, "unexpected status code %d", rr.Code)
}

func TestFileUploadWithNoFileAttachment(t *testing.T) {
	req := createFileUploadHttpRequest(t, "", someApiToken)
	rr := httptest.NewRecorder()
	createFileUploadHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %d", rr.Code)
}

func TestFileUploadDirMustBeAbsolute(t *testing.T) {
	handler := NewFileUploadHandler(".", someApiToken)
	test.Assertf(t, handler == nil, "expecting nil handler")
}

func TestFileUploadStoresFileUnderUuidDir(t *testing.T) {
	uploadDir, _ := ioutil.TempDir("", "uploads")
	handler := NewFileUploadHandler(uploadDir, someApiToken)
	req := createFileUploadHttpRequestTo(t, "testdata/some.txt", someApiToken)
	rr := httptest.NewRecorder()
	http.HandlerFunc(handler.Handle).ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusOK, "unexpected status code %d", rr.Code)

	var response FileUploadResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	test.Assertf(t, err == nil, "unable to unmarshal response : %v", err)
	test.Assert(t, response.Id != "", "expecting non-empty id")

	storedPath := filepath.Join(uploadDir, response.Id, "some.txt")
	_, err = os.Stat(storedPath)
	test.Assertf(t, err == nil, "expecting file to be stored at %s : %v", storedPath, err)
}

func createFileUploadHandler() http.Handler {
	uploadDir, _ := ioutil.TempDir("", "uploads")
	return http.HandlerFunc(NewFileUploadHandler(uploadDir, someApiToken).Handle)
}

func createFileUploadHttpRequest(t *testing.T, filename, apiToken string) *http.Request {
	return createFileUploadHttpRequestTo(t, filename, apiToken)
}

func createFileUploadHttpRequestTo(t *testing.T, filename, apiToken string) *http.Request {
	headers := map[string]string{}
	if apiToken != "" {
		headers["Authorization"] = "Bearer " + apiToken
	}
	req, err := uploader.CreateMultipartRequest("http://localhost", filename, map[string]string{}, headers)
	test.Assertf(t, err == nil, "unable to create request : %v", err)
	return req
}
