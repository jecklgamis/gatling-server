package handler

import (
	"os"

	"github.com/jecklgamis/gatling-server/pkg/api"
	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	"github.com/jecklgamis/gatling-server/pkg/jsonutil"
	"github.com/jecklgamis/gatling-server/pkg/s3"
	test "github.com/jecklgamis/gatling-server/pkg/testing"

	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func tempDir() string {
	dir, _ := os.MkdirTemp("", "")
	return dir
}

func tempFile(filename string) string {
	return filepath.Join(tempDir(), filename)
}

func TestApiHandlerWithInvalidMethod(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "http://localhost:8080", strings.NewReader("some-body"))
	createApiHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusMethodNotAllowed, "unexpected status code %d", rr.Code)
}

func TestApiHandlerWithNilBody(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "http://localhost:8080", nil)
	req.Header.Set("Authorization", "Bearer "+someApiToken)
	createApiHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %d", rr.Code)
}

func TestApiHandlerWithMissingApiToken(t *testing.T) {
	body := jsonutil.ToJson(createSubmitTaskRequest(
		"gatling.test.example.simulation.ExampleSimulation", "", "s3://some-bucket/gatling-scala-example-lean.jar"))
	req, err := http.NewRequest("POST", "/some-url", strings.NewReader(body))
	test.Assertf(t, err == nil, "unable to create request")
	rr := httptest.NewRecorder()
	createApiHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusUnauthorized, "unexpected status code %v", rr.Code)
}

func TestApiHandlerWithEmptySimulation(t *testing.T) {
	req := createSubmitTaskHttpRequest(t, "", "some-java-opts", "s3://some-bucket/some.jar")
	rr := httptest.NewRecorder()
	createApiHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %v", rr.Code)
}

func TestApiHandlerWithEmptyUrl(t *testing.T) {
	req := createSubmitTaskHttpRequest(t, "gatling.test.example.simulation.ExampleSimulation", "", "")
	rr := httptest.NewRecorder()
	createApiHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %v", rr.Code)
}

func TestApiHandlerWithInvalidJavaOpts(t *testing.T) {
	req := createSubmitTaskHttpRequest(t, "gatling.test.example.simulation.ExampleSimulation",
		"-javaagent:/tmp/evil.jar", "s3://some-bucket/some.jar")
	rr := httptest.NewRecorder()
	createApiHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %v", rr.Code)
}

func TestApiHandlerWithUnsupportedScheme(t *testing.T) {
	req := createSubmitTaskHttpRequest(t, "gatling.test.example.simulation.ExampleSimulation",
		"", "ftp://some-host/some.jar")
	rr := httptest.NewRecorder()
	createApiHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %v", rr.Code)
}

func TestApiHandlerS3NotEnabled(t *testing.T) {
	req := createSubmitTaskHttpRequest(t, "gatling.test.example.simulation.ExampleSimulation",
		"", "s3://some-bucket/some.jar")
	rr := httptest.NewRecorder()
	createApiHandlerWith(nil).ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %v", rr.Code)
}

func TestApiHandlerS3BucketNotAllowlisted(t *testing.T) {
	req := createSubmitTaskHttpRequest(t, "gatling.test.example.simulation.ExampleSimulation",
		"", "s3://some-other-bucket/some.jar")
	rr := httptest.NewRecorder()
	s3Ops := s3.NewFakeS3Ops(fileioutil.MustReadFile("testdata/gatling-scala-example-lean.jar"),
		tempFile("gatling-scala-example-lean.jar"), nil)
	createApiHandlerWith(s3Ops).ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %v", rr.Code)
}

func TestApiHandlerS3DownloadJarSimulation(t *testing.T) {
	req := createSubmitTaskHttpRequest(t,
		"gatling.test.example.simulation.ExampleSimulation",
		"-DbaseUrl=http://localhost:8080 -DdurationMin=0.5 -DrequestPersecond=1",
		"s3://some-bucket/gatling-scala-example-lean.jar")
	rr := httptest.NewRecorder()
	s3Ops := s3.NewFakeS3Ops(fileioutil.MustReadFile("testdata/gatling-scala-example-lean.jar"),
		tempFile("gatling-scala-example-lean.jar"), nil)
	createApiHandlerWith(s3Ops).ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusOK, "unexpected status code %v", rr.Code)
	validateSubmitTaskResponse(t, rr)
}

func TestApiHandlerHttpDownloadJarSimulation(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	defer ts.Close()
	req := createSubmitTaskHttpRequest(t,
		"gatling.test.example.simulation.ExampleSimulation",
		"-DbaseUrl=http://localhost:8080 -DdurationMin=0.5 -DrequestPersecond=1",
		ts.URL+"/gatling-scala-example-lean.jar")
	rr := httptest.NewRecorder()
	createApiHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusOK, "unexpected status code %v", rr.Code)
	validateSubmitTaskResponse(t, rr)
}

func TestApiHandlerHttpDownloadInvalidExtension(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	defer ts.Close()
	req := createSubmitTaskHttpRequest(t,
		"gatling.test.example.simulation.ExampleSimulation", "", ts.URL+"/some.txt")
	rr := httptest.NewRecorder()
	createApiHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %v", rr.Code)
}

func TestApiHandlerHttpDownloadNotFound(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	defer ts.Close()
	req := createSubmitTaskHttpRequest(t,
		"gatling.test.example.simulation.ExampleSimulation", "", ts.URL+"/does-not-exist.jar")
	rr := httptest.NewRecorder()
	createApiHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %v", rr.Code)
}

func createApiHandler() http.Handler {
	s3Ops := s3.NewFakeS3Ops(fileioutil.MustReadFile("testdata/gatling-scala-example-lean.jar"),
		tempFile("gatling-scala-example-lean.jar"), nil)
	return createApiHandlerWith(s3Ops)
}

func createApiHandlerWith(s3Ops s3.S3Ops) http.Handler {
	return http.HandlerFunc(NewApiHandler(someWorkspace(), someTaskManager(), s3Ops, someApiToken,
		nil, []string{"some-bucket"}, someApiToken, someApiToken).Handle)
}

func TestCheckHttpHostAllowedForDefaultAllowlistedHost(t *testing.T) {
	entry, err := checkHttpHostAllowed("localhost", DefaultAllowedHttpHosts)
	test.Assertf(t, err == nil, "expecting localhost to be allowed : %v", err)
	test.Assertf(t, entry != nil, "expecting localhost to be an explicit allowlist match")
}

func TestCheckHttpHostAllowedRejectsPrivateIP(t *testing.T) {
	_, err := checkHttpHostAllowed("10.0.0.5", DefaultAllowedHttpHosts)
	test.Assertf(t, err != nil, "expecting private IP to be rejected")
}

func TestCheckHttpHostAllowedRejectsLinkLocalMetadataIP(t *testing.T) {
	_, err := checkHttpHostAllowed("169.254.169.254", DefaultAllowedHttpHosts)
	test.Assertf(t, err != nil, "expecting link-local metadata IP to be rejected")
}

func TestCheckHttpHostAllowedAllowsPublicIP(t *testing.T) {
	entry, err := checkHttpHostAllowed("1.1.1.1", DefaultAllowedHttpHosts)
	test.Assertf(t, err == nil, "expecting public IP to be allowed : %v", err)
	test.Assertf(t, entry == nil, "expecting public IP fallback match to not return an allowlist entry")
}

func TestCheckHttpHostAllowedReturnsMatchedEntrysAuth(t *testing.T) {
	hosts := []AllowedHttpHost{{Host: "artifacts.example.com", Auth: &HttpHostAuth{Type: "bearer", Token: "some-token"}}}
	entry, err := checkHttpHostAllowed("artifacts.example.com", hosts)
	test.Assertf(t, err == nil, "expecting host to be allowed : %v", err)
	test.Assertf(t, entry != nil && entry.Auth != nil && entry.Auth.Token == "some-token",
		"expecting matched entry's auth to be returned")
}

func TestDownloadHttpFileAttachesBasicAuth(t *testing.T) {
	var gotUser, gotPass string
	var gotOk bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, gotOk = r.BasicAuth()
		w.Write([]byte("some-content"))
	}))
	defer ts.Close()
	_, err := downloadHttpFile(ts.URL+"/some.jar", tempDir(), &HttpHostAuth{Type: "basic", Username: "u", Password: "p"})
	test.Assertf(t, err == nil, "unexpected error : %v", err)
	test.Assertf(t, gotOk && gotUser == "u" && gotPass == "p", "expecting basic auth credentials to be sent")
}

func TestDownloadHttpFileAttachesBearerAuth(t *testing.T) {
	var gotAuthHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthHeader = r.Header.Get("Authorization")
		w.Write([]byte("some-content"))
	}))
	defer ts.Close()
	_, err := downloadHttpFile(ts.URL+"/some.jar", tempDir(), &HttpHostAuth{Type: "bearer", Token: "some-token"})
	test.Assertf(t, err == nil, "unexpected error : %v", err)
	test.Assertf(t, gotAuthHeader == "Bearer some-token", "unexpected Authorization header %q", gotAuthHeader)
}

func TestDownloadHttpFileSendsNoAuthWhenNil(t *testing.T) {
	var gotAuthHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthHeader = r.Header.Get("Authorization")
		w.Write([]byte("some-content"))
	}))
	defer ts.Close()
	_, err := downloadHttpFile(ts.URL+"/some.jar", tempDir(), nil)
	test.Assertf(t, err == nil, "unexpected error : %v", err)
	test.Assertf(t, gotAuthHeader == "", "expecting no Authorization header, got %q", gotAuthHeader)
}

func TestIsAllowedBucketMatch(t *testing.T) {
	test.Assertf(t, isAllowedBucket("some-bucket", []string{"some-bucket"}), "expecting bucket to be allowed")
}

func TestIsAllowedBucketNoMatch(t *testing.T) {
	test.Assertf(t, !isAllowedBucket("some-bucket", []string{"some-other-bucket"}), "expecting bucket to be denied")
}

func TestIsAllowedBucketEmptyListDeniesAll(t *testing.T) {
	test.Assertf(t, !isAllowedBucket("some-bucket", nil), "expecting empty allowlist to deny everything")
}

func createSubmitTaskHttpRequest(t *testing.T, simulation, javaOpts, url string) *http.Request {
	body := jsonutil.ToJson(createSubmitTaskRequest(simulation, javaOpts, url))
	req, err := http.NewRequest("POST", "/some-url", strings.NewReader(body))
	test.Assertf(t, err == nil, "unable to create request")
	req.Header.Set("Authorization", "Bearer "+someApiToken)
	return req
}

func createSubmitTaskRequest(simulation, javaOpts, url string) *api.SubmitTaskRequest {
	return &api.SubmitTaskRequest{
		Simulation: simulation,
		JavaOpts:   javaOpts,
		Url:        url}
}
