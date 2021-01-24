package handler

import (
	"encoding/json"
	"fmt"
	"github.com/jecklgamis/gatling-server/pkg/api"
	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	"github.com/jecklgamis/gatling-server/pkg/gatling"
	"github.com/jecklgamis/gatling-server/pkg/s3"
	"github.com/jecklgamis/gatling-server/pkg/tarutil"
	"github.com/jecklgamis/gatling-server/pkg/taskmanager"
	"github.com/jecklgamis/gatling-server/pkg/workspace"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

type S3DownloadHandler struct {
	WorkspaceOps workspace.Ops
	TaskOps      taskmanager.Ops
	s3Ops        s3.S3Ops
}

func NewS3DownloadHandler(workspaceOps workspace.Ops, taskOps taskmanager.Ops, s3Ops s3.S3Ops) *S3DownloadHandler {
	return &S3DownloadHandler{WorkspaceOps: workspaceOps, TaskOps: taskOps, s3Ops: s3Ops}
}

func (h *S3DownloadHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if r.Body == nil {
		badRequestWithError(w, fmt.Errorf("body is nil"))
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Unable to read request body :", err)
		internalServerError(w)
		return
	}
	request := api.S3DownloadTaskRequest{}
	if err := json.Unmarshal(body, &request); err != nil {
		log.Println("Unable to marshall request body :", err)
		badRequestWithError(w, fmt.Errorf("unable to marshall request body"))
		return
	}
	if err := validateRequest(&request); err != nil {
		log.Println("Invalid request :", err)
		badRequestWithError(w, err)
		return
	}
	taskId := taskmanager.CreateTaskId()
	taskPath := filepath.Join(h.WorkspaceOps.BaseDir(), taskId)
	userFilesDir, err := workspace.NewUserFilesDir(taskPath)
	if err != nil {
		log.Println("Unable to create user files directory :", err)
		internalServerError(w)
		return
	}
	_, _, err = s3.ParseS3Uri(request.Url)
	if err != nil {
		badRequest(w)
		return
	}
	storePath, err := h.s3Ops.DownloadUrl(request.Url, userFilesDir.BaseDir)
	if err != nil {
		log.Println("Unable to download file:", err)
		internalServerError(w)
		return
	}
	filename := filepath.Base(*storePath)
	if !hasValidFileExt(filename) {
		badRequestWithError(w, fmt.Errorf("invalid file extension"))
		return
	}
	task := gatling.NewTask(taskId, request.Simulation, request.JavaOpts, userFilesDir)
	if strings.HasSuffix(filename, ".scala") {
		log.Println("Submitting simulation", filename)
		destPath := fmt.Sprintf("%s/%s", userFilesDir.Simulations, filename)
		fileioutil.CopyFile(*storePath, destPath)
	} else {
		err := tarutil.Extract(*storePath, userFilesDir.BaseDir)
		if err != nil {
			log.Println("Unable to extract archive", err)
			badRequest(w)
			return
		}
	}
	metadata := &Metadata{TaskId: taskId, Simulation: request.Simulation, JavaOpts: request.JavaOpts}
	writeMetadata(userFilesDir.BaseDir, metadata, "metadata.json")
	_, err = h.TaskOps.SubmitTask(task)
	if err != nil {
		log.Println("Unable to submit task", err)
		internalServerError(w)
		return
	}
	okWithJson(w, &api.SubmitTaskResponse{Ok: true, TaskId: taskId})
}

func validateRequest(request *api.S3DownloadTaskRequest) error {
	if request.Simulation == "" {
		return fmt.Errorf("empty simulation class name")
	}
	return nil
}
