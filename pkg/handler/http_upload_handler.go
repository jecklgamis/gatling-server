package handler

import (
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
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const maxUploadSize = 500 << 20 // 500MB

type HttpUploadHandler struct {
	WorkspaceOps workspace.Ops
	TaskOps      taskmanager.Ops
	UploadDir    string
	ApiToken     string
	authLimiter  *authLimiter
}

type Metadata struct {
	TaskId     string `json:"taskId"`
	Simulation string `json:"simulation"`
	JavaOpts   string `json:"javaOpts"`
}

func NewHttpUploadHandler(workspace workspace.Ops, taskManager taskmanager.Ops, uploadDir string, apiToken string) *HttpUploadHandler {
	if !filepath.IsAbs(uploadDir) {
		slog.Error("Upload dir is not absolute")
		return nil
	}
	return &HttpUploadHandler{WorkspaceOps: workspace, TaskOps: taskManager, UploadDir: uploadDir, ApiToken: apiToken, authLimiter: newAuthLimiter()}
}

func (h *HttpUploadHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	clientKey := clientIP(r)
	if h.authLimiter.blocked(clientKey) {
		slog.Warn("Too many failed auth attempts from", "clientKey", clientKey)
		tooManyRequestsWithError(w, fmt.Errorf("too many failed authentication attempts"))
		return
	}
	if !isAuthorized(r, h.ApiToken) {
		h.authLimiter.recordFailure(clientKey)
		slog.Warn("Missing or invalid API token")
		unauthorizedWithError(w, fmt.Errorf("missing or invalid API token"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		slog.Error("Unable to parse multipart form", "error", err)
		badRequestWithError(w, fmt.Errorf("unable to parse multipart form"))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		slog.Error("No file uploaded", "error", err)
		badRequestWithError(w, fmt.Errorf("no file uploaded"))
		return
	}
	defer file.Close()

	filename := filepath.Base(header.Filename)
	if !hasValidFileExt(filename) {
		slog.Warn("Invalid file extension")
		badRequestWithError(w, fmt.Errorf("invalid file extension"))
		return
	}
	if err := validateFormFields(r); err != nil {
		slog.Warn("Missing required fields")
		badRequestWithError(w, err)
		return
	}
	storeDir := filepath.Join(h.UploadDir, uuid.New().String())
	err = fileioutil.CreateDirIfNotExist(storeDir, 0750)
	if err != nil {
		slog.Error("Unable to create dir", "error", err)
		internalServerError(w)
		return
	}
	defer func() {
		if err := os.RemoveAll(storeDir); err != nil {
			slog.Error("Unable to remove temporary upload dir", "error", err)
		}
	}()
	storePath := filepath.Join(storeDir, filename)
	if err := streamToFile(file, storePath); err != nil {
		slog.Error("Unable to store file", "error", err)
		internalServerError(w)
		return
	}
	taskId := taskmanager.CreateTaskId()
	taskPath := filepath.Join(h.WorkspaceOps.BaseDir(), taskId)
	userFilesDir, err := workspace.NewUserFilesDir(taskPath)
	if err != nil {
		slog.Error("Unable to create user files dir", "error", err)
		internalServerError(w)
		return
	}
	taskCommitted := false
	defer func() {
		if !taskCommitted {
			if err := os.RemoveAll(taskPath); err != nil {
				slog.Error("Unable to remove task dir after failed upload", "error", err)
			}
		}
	}()
	simulation := r.FormValue("simulation")
	javaOpts := r.FormValue("javaOpts")
	task := gatling.NewTask(taskId, simulation, javaOpts, userFilesDir)

	task.FileType = "jar"
	destPath := filepath.Join(userFilesDir.Simulations, filename)
	err = fileioutil.CopyFile(storePath, destPath)
	if err != nil {
		slog.Error("Unable to copy uploaded file to user files dir", "error", err)
		internalServerError(w)
		return
	}

	metadata := &Metadata{TaskId: taskId, Simulation: simulation, JavaOpts: javaOpts}
	err = writeMetadata(userFilesDir.BaseDir, metadata, "metadata.json")
	if err != nil {
		slog.Error("Unable write metadata file", "error", err)
		internalServerError(w)
		return
	}
	_, err = h.TaskOps.SubmitTask(task)
	if err != nil {
		slog.Error("Unable to submit task", "error", err)
		internalServerError(w)
		return
	}
	taskCommitted = true
	okWithJson(w, &api.SubmitTaskResponse{Ok: true, TaskId: taskId})
}

func writeMetadata(dir string, metadata *Metadata, filename string) error {
	path := filepath.Join(dir, filename)
	err := ioutil.WriteFile(path, []byte(jsonutil.ToJson(metadata)), 0640)
	if err != nil {
		slog.Error("Failed writing", "path", path, "error", err)
		return err
	}
	slog.Info("Wrote", "path", path)
	return nil
}

func streamToFile(src io.Reader, dst string) error {
	output, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer output.Close()
	_, err = io.Copy(output, src)
	return err
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
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
	if err := validateJavaOpts(r.FormValue("javaOpts")); err != nil {
		return err
	}
	return nil
}

// javaOptsTokenPattern restricts each whitespace-separated javaOpts token to a
// conservative set of characters that legitimate JVM flags are made of, ruling
// out shell metacharacters and anything else that could smuggle extra content
// into the token.
var javaOptsTokenPattern = regexp.MustCompile(`^-[A-Za-z0-9:=,._+/-]*$`)

// javaOptsForbiddenPrefixes blocks JVM flags that can be abused to execute
// arbitrary commands or load arbitrary native code (e.g. -XX:OnError and
// -XX:OnOutOfMemoryError run a shell command on crash/OOM; -javaagent,
// -agentlib and -agentpath load arbitrary code into the JVM).
var javaOptsForbiddenPrefixes = []string{
	"-xx:onerror",
	"-xx:onoutofmemoryerror",
	"-javaagent",
	"-agentlib",
	"-agentpath",
}

func validateJavaOpts(javaOpts string) error {
	for _, token := range strings.Fields(javaOpts) {
		if !javaOptsTokenPattern.MatchString(token) {
			return fmt.Errorf("invalid javaOpts token %q", token)
		}
		lower := strings.ToLower(token)
		for _, forbidden := range javaOptsForbiddenPrefixes {
			if strings.HasPrefix(lower, forbidden) {
				return fmt.Errorf("disallowed javaOpts flag %q", token)
			}
		}
	}
	return nil
}
