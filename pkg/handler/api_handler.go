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
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	maxApiTaskRequestSize = 1 << 20 // 1MB, this is a small JSON request body
	httpDownloadTimeout   = 5 * time.Minute
)

// ApiHandler accepts a generic task submission request whose Url may point to
// either an http(s) location or an s3:// location, downloads the referenced
// jar accordingly, and submits it as a task - unifying HttpUploadHandler's
// and S3DownloadHandler's submission flow behind a single endpoint.
type ApiHandler struct {
	WorkspaceOps workspace.Ops
	TaskOps      taskmanager.Ops
	s3Ops        s3.S3Ops
	ApiToken     string
	authLimiter  *authLimiter
}

func NewApiHandler(workspaceOps workspace.Ops, taskOps taskmanager.Ops, s3Ops s3.S3Ops, apiToken string) *ApiHandler {
	return &ApiHandler{WorkspaceOps: workspaceOps, TaskOps: taskOps, s3Ops: s3Ops, ApiToken: apiToken, authLimiter: newAuthLimiter()}
}

func (h *ApiHandler) Handle(w http.ResponseWriter, r *http.Request) {
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
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, maxApiTaskRequestSize+1))
	if err != nil {
		log.Println("Unable to read request body :", err)
		internalServerError(w)
		return
	}
	if len(body) > maxApiTaskRequestSize {
		badRequestWithError(w, fmt.Errorf("request body too large"))
		return
	}
	request := api.SubmitTaskRequest{}
	if err := json.Unmarshal(body, &request); err != nil {
		log.Println("Unable to marshall request body :", err)
		badRequestWithError(w, fmt.Errorf("unable to marshall request body"))
		return
	}
	if err := validateSubmitTaskRequest(&request); err != nil {
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

	storePath, err := h.download(request.Url, userFilesDir.BaseDir)
	if err != nil {
		log.Println("Unable to download file :", err)
		badRequestWithError(w, fmt.Errorf("unable to download file"))
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
	if err := fileioutil.CopyFile(*storePath, destPath); err != nil {
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
	if _, err := h.TaskOps.SubmitTask(task); err != nil {
		log.Println("Unable to submit task", err)
		internalServerError(w)
		return
	}
	taskCommitted = true
	okWithJson(w, &api.SubmitTaskResponse{Ok: true, TaskId: taskId})
}

// download fetches request.Url into dstDir, dispatching to the s3 or http(s)
// downloader based on the URL scheme, and returns the path it was stored at.
func (h *ApiHandler) download(rawUrl string, dstDir string) (*string, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(u.Scheme) {
	case "s3":
		if h.s3Ops == nil {
			return nil, fmt.Errorf("s3 downloads are not enabled")
		}
		return h.s3Ops.DownloadUrl(rawUrl, dstDir)
	case "http", "https":
		return downloadHttpFile(rawUrl, dstDir)
	default:
		return nil, fmt.Errorf("unsupported url scheme %q", u.Scheme)
	}
}

func downloadHttpFile(rawUrl string, dstDir string) (*string, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	filename := filepath.Base(u.Path)
	if filename == "" || filename == "." || filename == "/" {
		return nil, fmt.Errorf("unable to determine filename from url")
	}
	client := &http.Client{Timeout: httpDownloadTimeout}
	resp, err := client.Get(rawUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d downloading %s", resp.StatusCode, rawUrl)
	}
	storePath := filepath.Join(dstDir, filename)
	if err := streamToFile(io.LimitReader(resp.Body, maxUploadSize+1), storePath); err != nil {
		return nil, err
	}
	info, err := os.Stat(storePath)
	if err != nil {
		return nil, err
	}
	if info.Size() > maxUploadSize {
		os.Remove(storePath)
		return nil, fmt.Errorf("downloaded file too large")
	}
	return &storePath, nil
}

func validateSubmitTaskRequest(request *api.SubmitTaskRequest) error {
	if request.Simulation == "" {
		return fmt.Errorf("empty simulation class name")
	}
	if request.Url == "" {
		return fmt.Errorf("empty url")
	}
	if err := validateJavaOpts(request.JavaOpts); err != nil {
		return err
	}
	return nil
}
