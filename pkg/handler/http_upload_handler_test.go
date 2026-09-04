package handler

import (
	"encoding/json"
	"github.com/jecklgamis/gatling-server/pkg/api"
	"github.com/jecklgamis/gatling-server/pkg/gatling"
	"github.com/jecklgamis/gatling-server/pkg/taskmanager"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"github.com/jecklgamis/gatling-server/pkg/uploader"
	"github.com/jecklgamis/gatling-server/pkg/workspace"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUploadJarSimulation(t *testing.T) {
	rr := httptest.NewRecorder()
	createHandler().ServeHTTP(rr, someJarSimulationReq(t))
	validateSubmitTaskResponse(t, rr)
}

func TestUploadScalaFileRejected(t *testing.T) {
	req := createMultiPartRequest(t, "testdata/SingleFileExampleSimulation.scala",
		"gatling.test.example.simulation.SingleFileExampleSimulation", "")
	rr := httptest.NewRecorder()
	createHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %d", rr.Code)
}

func TestUploadTarGzFileRejected(t *testing.T) {
	req := createMultiPartRequest(t, "testdata/gatling-scala-example-user-files.tar.gz",
		"gatling.test.example.simulation.ExampleSimulation", "")
	rr := httptest.NewRecorder()
	createHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %d", rr.Code)
}

func TestFileUploadWithoutSimulationField(t *testing.T) {
	rr := httptest.NewRecorder()
	createHandler().ServeHTTP(rr, someMultipartRequestWithoutSimulationField(t))
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %d", rr.Code)
}

func TestUploadDirMustBeAbsolute(t *testing.T) {
	handler := NewHttpUploadHandler(someWorkspace(), someTaskManager(), ".", someApiToken)
	test.Assertf(t, handler == nil, "expecting nil handler")
}

func TestUploadWithMissingApiToken(t *testing.T) {
	req := createMultiPartRequestWithToken(t, "testdata/gatling-scala-example-lean.jar",
		"gatling.test.example.simulation.ExampleSimulation", "", "")
	rr := httptest.NewRecorder()
	createHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusUnauthorized, "unexpected status code %d", rr.Code)
}

func TestUploadWithInvalidApiToken(t *testing.T) {
	req := createMultiPartRequestWithToken(t, "testdata/gatling-scala-example-lean.jar",
		"gatling.test.example.simulation.ExampleSimulation", "", "some-invalid-token")
	rr := httptest.NewRecorder()
	createHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusUnauthorized, "unexpected status code %d", rr.Code)
}

func TestInvalidRequestMethod(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "http://localhost:8080", strings.NewReader("some-body"))
	createHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusMethodNotAllowed, "unexpected status code %d", rr.Code)
}

func TestUploadNilBody(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "http://localhost:8080", nil)
	req.Header.Set("Authorization", "Bearer "+someApiToken)
	createHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %d", rr.Code)
}

func TestNotMultipart(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "http://localhost:8080", strings.NewReader("some-body"))
	req.Header.Set("Authorization", "Bearer "+someApiToken)
	createHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %d", rr.Code)
}

func TestInvalidFileExtension(t *testing.T) {
	rr := httptest.NewRecorder()
	req := createMultiPartRequest(t, "testdata/some.txt", "some-simulation", "some-java-opts")
	createHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %d", rr.Code)
}

func TestNoFileAttachment(t *testing.T) {
	rr := httptest.NewRecorder()
	req := createMultiPartRequest(t, "", "some-simulation", "some-java-opts")
	createHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %d", rr.Code)
}

const someApiToken = "some-test-api-token"

func createHandler() http.Handler {
	uploadDir, _ := ioutil.TempDir("", "uploads")
	httpUploadHandler := NewHttpUploadHandler(someWorkspace(), someTaskManager(), uploadDir, someApiToken)
	return http.HandlerFunc(httpUploadHandler.Handle)
}

func someTaskManager() *taskmanager.TaskManager {
	return taskmanager.NewTaskManager(gatling.SomeGatling(), make(chan interface{}, 1024), []uploader.GatlingArtifactUploader{})
}

func someWorkspace() *workspace.Workspace {
	dir, _ := ioutil.TempDir("", "repos")
	return workspace.NewWorkspace(dir)
}

func someJarSimulationReq(t *testing.T) *http.Request {
	return createMultiPartRequest(t, "testdata/gatling-scala-example-lean.jar",
		"gatling.test.example.simulation.ExampleSimulation",
		"-DbaseUrl=http://localhost:8080 -DdurationMin=0.10 -DrequestPersecond=1")
}

func someMultipartRequestWithoutSimulationField(t *testing.T) *http.Request {
	return createMultiPartRequest(t, "testdata/gatling-scala-example-lean.jar", "",
		"-DsomeKey=someValue")
}

func createMultiPartRequest(t *testing.T, filename, simulation, javaOpts string) *http.Request {
	return createMultiPartRequestWithToken(t, filename, simulation, javaOpts, someApiToken)
}

func createMultiPartRequestWithToken(t *testing.T, filename, simulation, javaOpts, apiToken string) *http.Request {
	kv := map[string]string{}
	if simulation != "" {
		kv["simulation"] = simulation
	}
	if javaOpts != "" {
		kv["javaOpts"] = javaOpts
	}
	headers := map[string]string{}
	if apiToken != "" {
		headers["Authorization"] = "Bearer " + apiToken
	}
	req, err := uploader.CreateMultipartRequest("http://localhost", filename, kv, headers)
	test.Assertf(t, err == nil, "unable create request : %v", err)
	return req
}

func validateSubmitTaskResponse(t *testing.T, rr *httptest.ResponseRecorder) {
	test.Assertf(t, rr.Code == http.StatusOK, "unexpected status code %d", rr.Code)
	test.Assertf(t, rr.Header().Get("Content-Type") == "application/json",
		"unexpected content type %s", rr.Header().Get("Content-Type"))
	var entity api.SubmitTaskResponse
	json.Unmarshal(rr.Body.Bytes(), &entity)
	test.Assert(t, entity.Ok, "expecting ok result")
	test.Assert(t, entity.TaskId != "", "task id is empty")
}
