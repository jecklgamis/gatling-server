package handler

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"github.com/google/uuid"
	"github.com/jecklgamis/gatling-server/pkg/api"
	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	"github.com/jecklgamis/gatling-server/pkg/gatling"
	"github.com/jecklgamis/gatling-server/pkg/jsonutil"
	"github.com/jecklgamis/gatling-server/pkg/taskmanager"
	"github.com/jecklgamis/gatling-server/pkg/workspace"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const maxUploadSize = 500 << 20 // 500MB

type HttpUploadHandler struct {
	WorkspaceOps workspace.Ops
	TaskOps      taskmanager.Ops
	UploadDir    string
	ApiToken     string
}

type Metadata struct {
	TaskId     string `json:"taskId"`
	Simulation string `json:"simulation"`
	JavaOpts   string `json:"javaOpts"`
}

func NewHttpUploadHandler(workspace workspace.Ops, taskManager taskmanager.Ops, uploadDir string, apiToken string) *HttpUploadHandler {
	if !filepath.IsAbs(uploadDir) {
		log.Println("Upload dir is not absolute")
		return nil
	}
	return &HttpUploadHandler{WorkspaceOps: workspace, TaskOps: taskManager, UploadDir: uploadDir, ApiToken: apiToken}
}

func (h *HttpUploadHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !isAuthorized(r, h.ApiToken) {
		log.Println("Missing or invalid API token")
		unauthorizedWithError(w, fmt.Errorf("missing or invalid API token"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Println("Unable to parse multipart form :", err)
		badRequestWithError(w, fmt.Errorf("unable to parse multipart form"))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Println("No file uploaded :", err)
		badRequestWithError(w, fmt.Errorf("no file uploaded"))
		return
	}
	defer file.Close()

	filename := filepath.Base(header.Filename)
	if !hasValidFileExt(filename) {
		log.Println("Invalid file extension")
		badRequestWithError(w, fmt.Errorf("invalid file extension"))
		return
	}
	if err := validateFormFields(r); err != nil {
		log.Println("Missing required fields")
		badRequestWithError(w, err)
		return
	}
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, file)
	if err != nil {
		log.Println("Unable to copy file :", err)
		internalServerError(w)
		return
	}
	storeDir := filepath.Join(h.UploadDir, uuid.New().String()[0:8])
	err = fileioutil.CreateDirIfNotExist(storeDir, 0744)
	if err != nil {
		log.Println("Unable to create dir :", err)
		internalServerError(w)
		return
	}
	defer func() {
		if err := os.RemoveAll(storeDir); err != nil {
			log.Println("Unable to remove temporary upload dir :", err)
		}
	}()
	storePath, err := fileioutil.WriteBufferToFile(&buffer, storeDir, filename)
	if err != nil {
		log.Println("Unable to store file :", err)
		internalServerError(w)
		return
	}
	taskId := taskmanager.CreateTaskId()
	taskPath := filepath.Join(h.WorkspaceOps.BaseDir(), taskId)
	userFilesDir, err := workspace.NewUserFilesDir(taskPath)
	if err != nil {
		log.Println("Unable to create user files dir :", err)
		internalServerError(w)
		return
	}
	taskCommitted := false
	defer func() {
		if !taskCommitted {
			if err := os.RemoveAll(taskPath); err != nil {
				log.Println("Unable to remove task dir after failed upload :", err)
			}
		}
	}()
	simulation := r.FormValue("simulation")
	javaOpts := r.FormValue("javaOpts")
	task := gatling.NewTask(taskId, simulation, javaOpts, userFilesDir)

	task.FileType = "jar"
	destPath := filepath.Join(userFilesDir.Simulations, filename)
	err = fileioutil.CopyFile(*storePath, destPath)
	if err != nil {
		log.Println("Unable to copy uploaded file to user files dir : ", err)
		internalServerError(w)
		return
	}

	metadata := &Metadata{TaskId: taskId, Simulation: simulation, JavaOpts: javaOpts}
	err = writeMetadata(userFilesDir.BaseDir, metadata, "metadata.json")
	if err != nil {
		log.Println("Unable write metadata file :", err)
		internalServerError(w)
		return
	}
	_, err = h.TaskOps.SubmitTask(task)
	if err != nil {
		log.Println("Unable to submit task :", err)
		internalServerError(w)
		return
	}
	taskCommitted = true
	okWithJson(w, &api.SubmitTaskResponse{Ok: true, TaskId: taskId})
}

func writeMetadata(dir string, metadata *Metadata, filename string) error {
	path := filepath.Join(dir, filename)
	err := ioutil.WriteFile(path, []byte(jsonutil.ToJson(metadata)), 0744)
	if err != nil {
		log.Println("Failed writing", path)
		return err
	}
	log.Println("Wrote", path)
	return nil
}

func hasValidFileExt(filename string) bool {
	return strings.HasSuffix(filename, ".jar")
}

func isAuthorized(r *http.Request, apiToken string) bool {
	if apiToken == "" {
		return false
	}
	const prefix = "Bearer "
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, prefix) {
		return false
	}
	token := strings.TrimPrefix(authHeader, prefix)
	return subtle.ConstantTimeCompare([]byte(token), []byte(apiToken)) == 1
}

func validateFormFields(r *http.Request) error {
	if r.FormValue("simulation") == "" {
		return fmt.Errorf("expecting simulation key")
	}
	return nil
}
