package handler

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	"github.com/jecklgamis/gatling-server/pkg/taskmanager"
	"github.com/jecklgamis/gatling-server/pkg/workspace"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
)

type TaskHandler struct {
	WorkspaceOps workspace.Ops
	TaskOps      taskmanager.Ops
	UploadDir    string
}

func NewTaskHandler(workspace workspace.Ops, taskOps taskmanager.Ops) *TaskHandler {
	return &TaskHandler{WorkspaceOps: workspace, TaskOps: taskOps}
}

func (h *TaskHandler) MetadataHandler(w http.ResponseWriter, r *http.Request) {
	h.serveFileFromWorkspace(w, r, "metadata.json", "application/json")
}

func (h *TaskHandler) ResultsHandler(w http.ResponseWriter, r *http.Request) {
	h.serveFileFromWorkspace(w, r, "results.tar.gz", "application/octet-stream")
}

func (h *TaskHandler) ConsoleLogHandler(w http.ResponseWriter, r *http.Request) {
	h.serveFileFromWorkspace(w, r, "console.log", "text/plain")
}

func (h *TaskHandler) TaskContextHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId := vars["taskId"]
	taskContext, found := h.TaskOps.GetTaskRuntimeContext(taskId)
	if !found {
		notFound(w)
		return
	}
	okWithJson(w, taskContext)
}

func (h *TaskHandler) AbortTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId := vars["taskId"]
	_, found := h.TaskOps.GetTaskRuntimeContext(taskId)
	if !found {
		notFound(w)
		return
	}
	err := h.TaskOps.AbortTask(taskId)
	if err != nil {
		log.Println("Unable to abort task :", err)
		internalServerError(w)
		return
	}
	ok(w)
}

func (h *TaskHandler) SimulationLogHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId := vars["taskId"]
	if taskId == "" {
		badRequestWithError(w, fmt.Errorf("requires task id"))
		return
	}
	dir := filepath.Join(h.WorkspaceOps.BaseDir(), taskId, "results")
	if !fileioutil.DirExists(dir) {
		notFound(w)
		return
	}
	simulationLog, err := fileioutil.FindFile(dir, "simulation.log")
	if err != nil {
		notFound(w)
		return
	}
	bytes, err := ioutil.ReadFile(simulationLog)
	if err != nil {
		notFound(w)
		return
	}
	okWithEntity(w, "text/plain", bytes)
}

func (h *TaskHandler) serveFileFromWorkspace(w http.ResponseWriter, r *http.Request, file string, contentType string) {
	vars := mux.Vars(r)
	taskId := vars["taskId"]
	if taskId == "" {
		badRequestWithError(w, fmt.Errorf("requires task id"))
		return
	}
	content, err := h.WorkspaceOps.ReadFile(taskId, file)
	if err != nil {
		notFound(w)
		return
	}
	okWithEntity(w, contentType, content)
}
