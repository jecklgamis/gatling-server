package handler

import (
	"github.com/jecklgamis/gatling-server/pkg/api"
	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	"github.com/jecklgamis/gatling-server/pkg/jsonutil"
	"github.com/jecklgamis/gatling-server/pkg/s3"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestS3HandlerWithInvalidMethod(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "http://localhost:8080", strings.NewReader("some-body"))
	createS3Handler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusMethodNotAllowed, "unexpected status code %d", rr.Code)
}

func TestS3HandlerWithNilBody(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "http://localhost:8080", nil)
	createS3Handler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %d", rr.Code)
}

func TestS3HandlerWithEmptySimulation(t *testing.T) {
	req := createS3DownloadHttpRequest(t, "", "some-java-opts", "some-s3-url")
	rr := httptest.NewRecorder()
	createS3Handler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %v", rr.Code)
}

func TestS3HandlerWithInvalidDownloadedFileExtension(t *testing.T) {
	req := createS3DownloadHttpRequest(t,
		"gatling.test.example.simulation.SingleFileExampleSimulation",
		"-DbaseUrl=http://localhost:8080 -DdurationMin=0.5 -DrequestPersecond=1",
		"s3://some-bucket/SingleFileExampleSimulation.scala")
	rr := httptest.NewRecorder()
	s3Ops := s3.NewFakeS3Ops(fileioutil.MustReadFile("testdata/SingleFileExampleSimulation.scala"),
		tempFile("ExampleSimulation.blah"), nil)
	createS3HandlerWith(s3Ops).ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %v", rr.Code)
}

func TestDownloadSingFileSimulation(t *testing.T) {
	req := createS3DownloadHttpRequest(t,
		"gatling.test.example.simulation.SingleFileExampleSimulation",
		"-DbaseUrl=http://localhost:8080 -DdurationMin=0.5 -DrequestPersecond=1",
		"s3://some-bucket/SingleFileExampleSimulation.scala")
	rr := httptest.NewRecorder()
	s3Ops := s3.NewFakeS3Ops(fileioutil.MustReadFile("testdata/SingleFileExampleSimulation.scala"),
		tempFile("SingleFileExampleSimulation.scala"), nil)
	createS3HandlerWith(s3Ops).ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusOK, "unexpected status code %v", rr.Code)
	validateSubmitTaskResponse(t, rr)
}

func TestDownloadPackagedSingleFileSimulation(t *testing.T) {
	req := createS3DownloadHttpRequest(t,
		"gatling.test.example.simulation.ExampleSimulation",
		"-DbaseUrl=http://localhost:8080 -DdurationMin=0.5 -DrequestPersecond=1",
		"s3://some-bucket/gatling-test-example-user-files.tar.gz")
	rr := httptest.NewRecorder()
	s3Ops := s3.NewFakeS3Ops(fileioutil.MustReadFile("testdata/gatling-test-example-user-files.tar.gz"),
		tempFile("gatling-test-example-user-files.tar.gz"), nil)
	createS3HandlerWith(s3Ops).ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusOK, "unexpected status code %v", rr.Code)
	validateSubmitTaskResponse(t, rr)
}

func TestDownloadCorruptedTarGz(t *testing.T) {
	req := createS3DownloadHttpRequest(t,
		"gatling.test.example.simulation.ExampleSimulation",
		"-DbaseUrl=http://localhost:8080 -DdurationMin=0.5 -DrequestPersecond=1",
		"s3://some-bucket/gatling-test-example-user-files.tar.gz")
	rr := httptest.NewRecorder()
	s3Ops := s3.NewFakeS3Ops(fileioutil.MustReadFile("testdata/corrupted.tar.gz"),
		tempFile("corrupted.tar.gz"), nil)
	createS3HandlerWith(s3Ops).ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %v", rr.Code)
}

func createS3Handler() http.Handler {
	s3Ops := s3.NewFakeS3Ops(fileioutil.MustReadFile("testdata/SingleFileExampleSimulation.scala"),
		tempFile("ExampleSimulation.scala"), nil)
	return http.HandlerFunc(NewS3DownloadHandler(someWorkspace(), someTaskManager(), s3Ops).Handle)
}

func createS3HandlerWith(s3Ops s3.S3Ops) http.Handler {
	return http.HandlerFunc(NewS3DownloadHandler(someWorkspace(), someTaskManager(), s3Ops).Handle)
}

func tempDir() string {
	dir, _ := ioutil.TempDir("", "")
	return dir
}

func tempFile(filename string) string {
	return filepath.Join(tempDir(), filename)
}

func createS3DownloadHttpRequest(t *testing.T, simulation, javaOpts, s3URL string) *http.Request {
	body := jsonutil.ToJson(createS3DownloadRequest(simulation, javaOpts, s3URL))
	req, err := http.NewRequest("POST", "/some-url", strings.NewReader(body))
	test.Assertf(t, err == nil, "unable to create request")
	return req
}

func createS3DownloadRequest(simulation, javaOpts, s3URL string) *api.S3DownloadTaskRequest {
	return &api.S3DownloadTaskRequest{
		Simulation: simulation,
		JavaOpts:   javaOpts,
		Url:        s3URL}
}
