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

func TestUploadSingleFileSimulation(t *testing.T) {
	rr := httptest.NewRecorder()
	createHandler().ServeHTTP(rr, someSingleFileSimulationReq(t))
	validateSubmitTaskResponse(t, rr)
}

func TestUploadPackagedSimulation(t *testing.T) {
	rr := httptest.NewRecorder()
	createHandler().ServeHTTP(rr, somePackagedSimulationReq(t))
	validateSubmitTaskResponse(t, rr)
}

func TestUploadFileWithPackagedSimulation(t *testing.T) {
	req := createMultiPartRequest(t, "testdata/gatling-test-example-user-files.tar.gz",
		"gatling.test.example.simulation.SingleFileExampleSimulation", "")
	rr := httptest.NewRecorder()
	createHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusOK, "unexpected status code %d", rr.Code)
}

func TestFileUploadWithoutSimulationField(t *testing.T) {
	rr := httptest.NewRecorder()
	createHandler().ServeHTTP(rr, someMultipartRequestWithoutSimulationField(t))
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %d", rr.Code)
}

func TestUploadDirMustBeAbsolute(t *testing.T) {
	handler := NewHttpUploadHandler(someWorkspace(), someTaskManager(), ".")
	test.Assertf(t, handler == nil, "expecting nil handler")
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
	createHandler().ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusBadRequest, "unexpected status code %d", rr.Code)
}

func TestNotMultipart(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "http://localhost:8080", strings.NewReader("some-body"))
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

func createHandler() http.Handler {
	uploadDir, _ := ioutil.TempDir("", "uploads")
	httpUploadHandler := NewHttpUploadHandler(someWorkspace(), someTaskManager(), uploadDir)
	return http.HandlerFunc(httpUploadHandler.Handle)
}

func someTaskManager() *taskmanager.TaskManager {
	return taskmanager.NewTaskManager(gatling.SomeGatlingDist(), make(chan interface{}, 1024), []uploader.GatlingArtifactUploader{})
}

func someWorkspace() *workspace.Workspace {
	dir, _ := ioutil.TempDir("", "workspace")
	return workspace.NewWorkspace(dir)
}

func someSingleFileSimulationReq(t *testing.T) *http.Request {
	return createMultiPartRequest(t, "testdata/SingleFileExampleSimulation.scala",
		"gatling.test.example.simulation.SingleFileExampleSimulation",
		"-DbaseUrl=http://localhost:8080 -DdurationMin=0.10 -DrequestPersecond=1")
}

func somePackagedSimulationReq(t *testing.T) *http.Request {
	return createMultiPartRequest(t, "testdata/gatling-test-example-user-files.tar.gz",
		"gatling.test.example.simulation.ExampleSimulation",
		"-DbaseUrl=http://localhost:8080 -DdurationMin=0.10 -DrequestPersecond=1")
}

func someMultipartRequestWithoutSimulationField(t *testing.T) *http.Request {
	return createMultiPartRequest(t, "testdata/SingleFileExampleSimulation.scala", "",
		"-DsomeKey=someValue")
}

func createMultiPartRequest(t *testing.T, filename, simulation, javaOpts string) *http.Request {
	kv := map[string]string{}
	if simulation != "" {
		kv["simulation"] = simulation
	}
	if javaOpts != "" {
		kv["javaOpts"] = javaOpts
	}
	req, err := uploader.CreateMultipartRequest("http://localhost", filename, kv)
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
