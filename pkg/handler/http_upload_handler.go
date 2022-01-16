package handler

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/jecklgamis/gatling-server/pkg/api"
	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	"github.com/jecklgamis/gatling-server/pkg/gatling"
	"github.com/jecklgamis/gatling-server/pkg/jsonutil"
	"github.com/jecklgamis/gatling-server/pkg/tarutil"
	"github.com/jecklgamis/gatling-server/pkg/taskmanager"
	"github.com/jecklgamis/gatling-server/pkg/workspace"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

type HttpUploadHandler struct {
	WorkspaceOps workspace.Ops
	TaskOps      taskmanager.Ops
	UploadDir    string
}

type Metadata struct {
	TaskId     string `json:"taskId"`
	Simulation string `json:"simulation"`
	JavaOpts   string `json:"javaOpts"`
}

func NewHttpUploadHandler(workspace workspace.Ops, taskManager taskmanager.Ops, uploadDir string) *HttpUploadHandler {
	if !filepath.IsAbs(uploadDir) {
		log.Println("Upload dir is not absolute")
		return nil
	}
	return &HttpUploadHandler{WorkspaceOps: workspace, TaskOps: taskManager, UploadDir: uploadDir}
}

func (h *HttpUploadHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
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

	if !hasValidFileExt(header.Filename) {
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
	}
	storePath, err := fileioutil.WriteBufferToFile(&buffer, storeDir, header.Filename)
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
	simulation := r.FormValue("simulation")
	javaOpts := r.FormValue("javaOpts")
	task := gatling.NewTask(taskId, simulation, javaOpts, userFilesDir)

	if strings.HasSuffix(header.Filename, ".scala") {
		task.FileType = "scala"
		destPath := filepath.Join(userFilesDir.Simulations, header.Filename)
		err := fileioutil.CopyFile(*storePath, destPath)
		if err != nil {
			log.Println("Unable to copy uploaded file to user files dir : ", err)
			internalServerError(w)
			return
		}
	} else if strings.HasSuffix(header.Filename, ".jar") {
		task.FileType = "jar"
		destPath := filepath.Join(userFilesDir.Simulations, header.Filename)
		err := fileioutil.CopyFile(*storePath, destPath)
		if err != nil {
			log.Println("Unable to copy uploaded file to user files dir : ", err)
			internalServerError(w)
			return
		}
	} else if strings.HasSuffix(header.Filename, ".tar.gz") {
		task.FileType = "tar.gz"
		err := tarutil.Extract(*storePath, userFilesDir.BaseDir)
		if err != nil {
			log.Println("Unable to extract archive file :", err)
			internalServerError(w)
			return
		}
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
	return strings.HasSuffix(filename, ".scala") || strings.HasSuffix(filename, ".jar") ||
		strings.HasSuffix(filename, ".tar.gz")
}

func validateFormFields(r *http.Request) error {
	if r.FormValue("simulation") == "" {
		return fmt.Errorf("expecting simulation key")
	}
	return nil
}
