package handler

import (
	"github.com/gorilla/mux"
	"github.com/jecklgamis/gatling-server/pkg/gatling"
	"github.com/jecklgamis/gatling-server/pkg/taskmanager"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"github.com/jecklgamis/gatling-server/pkg/workspace"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetMetadata(t *testing.T) {
	rr := httptest.NewRecorder()
	req := someGetRequest(t, "/task/metadata/some-task-id")
	mux.SetURLVars(req, map[string]string{"taskId": "some-task-id"})

	workspaceOps := someWorkspace()
	handler := http.HandlerFunc(NewTaskHandler(workspaceOps, someTaskManager()).MetadataHandler)
	router := mux.NewRouter()
	router.HandleFunc("/task/metadata/{taskId}", handler)

	creteSomeUsersFilesDir(t, workspaceOps, "some-task-id")
	router.ServeHTTP(rr, req)
	validateOk(t, rr, "application/json")
}

func TestGetConsoleLog(t *testing.T) {
	rr := httptest.NewRecorder()
	req := someGetRequest(t, "/task/console/some-task-id")
	mux.SetURLVars(req, map[string]string{"taskId": "some-task-id"})

	workspaceOps := someWorkspace()
	handler := http.HandlerFunc(NewTaskHandler(workspaceOps, someTaskManager()).ConsoleLogHandler)
	router := mux.NewRouter()
	router.HandleFunc("/task/console/{taskId}", handler)

	creteSomeUsersFilesDir(t, workspaceOps, "some-task-id")
	router.ServeHTTP(rr, req)
	validateOk(t, rr, "text/plain")
}

func TestSimulationLog(t *testing.T) {
	rr := httptest.NewRecorder()
	req := someGetRequest(t, "/task/console/some-task-id")
	mux.SetURLVars(req, map[string]string{"taskId": "some-task-id"})

	workspaceOps := someWorkspace()
	handler := http.HandlerFunc(NewTaskHandler(workspaceOps, someTaskManager()).SimulationLogHandler)
	router := mux.NewRouter()
	router.HandleFunc("/task/console/{taskId}", handler)

	creteSomeUsersFilesDir(t, workspaceOps, "some-task-id")
	router.ServeHTTP(rr, req)
	validateOk(t, rr, "text/plain")
}

func TestGetResults(t *testing.T) {
	rr := httptest.NewRecorder()
	req := someGetRequest(t, "/task/results/some-task-id")
	mux.SetURLVars(req, map[string]string{"taskId": "some-task-id"})

	workspaceOps := someWorkspace()
	handler := http.HandlerFunc(NewTaskHandler(workspaceOps, someTaskManager()).ResultsHandler)
	router := mux.NewRouter()
	router.HandleFunc("/task/results/{taskId}", handler)

	creteSomeUsersFilesDir(t, workspaceOps, "some-task-id")
	router.ServeHTTP(rr, req)
	validateOk(t, rr, "application/octet-stream")
}

func TestGetMetadataWithoutTaskId(t *testing.T) {
	taskHandler := NewTaskHandler(someWorkspace(), someTaskManager())
	handler := http.HandlerFunc(taskHandler.MetadataHandler)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, someGetRequest(t, ""))
	test.Assertf(t, rr.Code == http.StatusBadRequest, "expecting 400")
}

func TestAbortTask(t *testing.T) {
	rr := httptest.NewRecorder()
	req := somePostRequest(t, "/task/abort/some-task-id")
	mux.SetURLVars(req, map[string]string{"taskId": "some-task-id"})

	workspaceOps := someWorkspace()
	taskManager := someTaskManager()
	taskManager.TaskContexts["some-task-id"] = &taskmanager.TaskRuntimeContext{
		Process: someProcess(), Task: &gatling.Task{Id: "some-task-id"}}
	handler := http.HandlerFunc(NewTaskHandler(workspaceOps, taskManager).AbortTaskHandler)
	router := mux.NewRouter()
	router.HandleFunc("/task/abort/{taskId}", handler)

	creteSomeUsersFilesDir(t, workspaceOps, "some-task-id")
	router.ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusOK, "expecting 200 but got %d", rr.Code)
}

func TestAbortUnknownTask(t *testing.T) {
	rr := httptest.NewRecorder()
	req := somePostRequest(t, "/task/abort/some-task-id")
	mux.SetURLVars(req, map[string]string{"taskId": "some-task-id"})

	workspaceOps := someWorkspace()
	taskManager := someTaskManager()
	handler := http.HandlerFunc(NewTaskHandler(workspaceOps, taskManager).AbortTaskHandler)
	router := mux.NewRouter()
	router.HandleFunc("/task/abort/{taskId}", handler)

	creteSomeUsersFilesDir(t, workspaceOps, "some-task-id")
	router.ServeHTTP(rr, req)
	test.Assertf(t, rr.Code == http.StatusNotFound, "expecting 400 but got %d", rr.Code)
}

func someProcess() *os.Process {
	cmd := exec.Command("ls")
	cmd.Start()
	return cmd.Process
}

func creteSomeUsersFilesDir(t *testing.T, workspaceOps workspace.Ops, id string) *workspace.UserFilesDir {
	userFilesDir, err := workspaceOps.NewUserFilesDir(id)
	test.Assertf(t, err == nil, "unable to create user files dir")

	file := filepath.Join(userFilesDir.BaseDir, "console.log")
	content := []byte("some log output")
	err = ioutil.WriteFile(file, content, 0744)
	test.Assertf(t, err == nil, "unable to create file :%v", err)

	file = filepath.Join(userFilesDir.Results, "simulation.log")
	content = []byte("some simulation output")
	err = ioutil.WriteFile(file, content, 0744)
	test.Assertf(t, err == nil, "unable to create file :%v", err)

	file = filepath.Join(userFilesDir.BaseDir, "metadata.json")
	content = []byte(`{"some-key":"some-value"}`)
	err = ioutil.WriteFile(file, content, 0744)
	test.Assertf(t, err == nil, "unable to create file :%v", err)

	file = filepath.Join(userFilesDir.BaseDir, "results.tar.gz")
	content = []byte(`some-tar-content"}`)
	err = ioutil.WriteFile(file, content, 0744)
	test.Assertf(t, err == nil, "unable to create file :%v", err)

	return userFilesDir
}

func validateOk(t *testing.T, rr *httptest.ResponseRecorder, contentType string) {
	test.Assertf(t, rr.Code == http.StatusOK, "unexpected status code %d", rr.Code)
	test.Assertf(t, strings.Contains(rr.Header().Get("Content-Type"), contentType), "unexpected content type :%v",
		rr.Header().Get("Content-Type"))
}

func someGetRequest(t *testing.T, path string) *http.Request {
	request, err := http.NewRequest("GET", path, nil)
	test.Assertf(t, err == nil, "failed to create request")
	return request
}

func somePostRequest(t *testing.T, path string) *http.Request {
	request, err := http.NewRequest("POST", path, nil)
	test.Assertf(t, err == nil, "failed to create request")
	return request
}
