package handler

import (
	"encoding/json"
	"fmt"
	"github.com/jecklgamis/gatling-server/pkg/api"
	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	"github.com/jecklgamis/gatling-server/pkg/gatling"
	"github.com/jecklgamis/gatling-server/pkg/s3"
	"github.com/jecklgamis/gatling-server/pkg/taskmanager"
	"github.com/jecklgamis/gatling-server/pkg/workspace"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const maxS3DownloadRequestSize = 1 << 20 // 1MB, this is a small JSON request body

type S3DownloadHandler struct {
	WorkspaceOps workspace.Ops
	TaskOps      taskmanager.Ops
	s3Ops        s3.S3Ops
	ApiToken     string
	authLimiter  *authLimiter
}

func NewS3DownloadHandler(workspaceOps workspace.Ops, taskOps taskmanager.Ops, s3Ops s3.S3Ops, apiToken string) *S3DownloadHandler {
	return &S3DownloadHandler{WorkspaceOps: workspaceOps, TaskOps: taskOps, s3Ops: s3Ops, ApiToken: apiToken, authLimiter: newAuthLimiter()}
}

func (h *S3DownloadHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	clientKey := clientIP(r)
	if h.authLimiter.blocked(clientKey) {
		log.Println("Too many failed auth attempts from", clientKey)
		tooManyRequestsWithError(w, fmt.Errorf("too many failed authentication attempts"))
		return
	}
	if !isAuthorized(r, h.ApiToken) {
		h.authLimiter.recordFailure(clientKey)
		log.Println("Missing or invalid API token")
		unauthorizedWithError(w, fmt.Errorf("missing or invalid API token"))
		return
	}
	if r.Body == nil {
		badRequestWithError(w, fmt.Errorf("body is nil"))
		return
	}
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, maxS3DownloadRequestSize+1))
	if err != nil {
		log.Println("Unable to read request body :", err)
		internalServerError(w)
		return
	}
	if len(body) > maxS3DownloadRequestSize {
		badRequestWithError(w, fmt.Errorf("request body too large"))
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
	taskCommitted := false
	defer func() {
		if !taskCommitted {
			if err := os.RemoveAll(taskPath); err != nil {
				log.Println("Unable to remove task dir after failed download :", err)
			}
		}
	}()
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
	task.FileType = "jar"
	log.Println("Submitting simulation", filename)
	destPath := filepath.Join(userFilesDir.Simulations, filename)
	err = fileioutil.CopyFile(*storePath, destPath)
	if err != nil {
		log.Println("Unable to copy downloaded file to user files dir :", err)
		internalServerError(w)
		return
	}
	metadata := &Metadata{TaskId: taskId, Simulation: request.Simulation, JavaOpts: request.JavaOpts}
	if err := writeMetadata(userFilesDir.BaseDir, metadata, "metadata.json"); err != nil {
		log.Println("Unable write metadata file :", err)
		internalServerError(w)
		return
	}
	_, err = h.TaskOps.SubmitTask(task)
	if err != nil {
		log.Println("Unable to submit task", err)
		internalServerError(w)
		return
	}
	taskCommitted = true
	okWithJson(w, &api.SubmitTaskResponse{Ok: true, TaskId: taskId})
}

func validateRequest(request *api.S3DownloadTaskRequest) error {
	if request.Simulation == "" {
		return fmt.Errorf("empty simulation class name")
	}
	if err := validateJavaOpts(request.JavaOpts); err != nil {
		return err
	}
	return nil
}
