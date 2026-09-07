package handler

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	"log/slog"
	"net/http"
	"path/filepath"
)

type FileUploadHandler struct {
	UploadDir   string
	ApiToken    string
	authLimiter *authLimiter
}

type FileUploadResponse struct {
	Id string `json:"id"`
}

func NewFileUploadHandler(uploadDir string, apiToken string) *FileUploadHandler {
	if !filepath.IsAbs(uploadDir) {
		slog.Error("Upload dir is not absolute")
		return nil
	}
	return &FileUploadHandler{UploadDir: uploadDir, ApiToken: apiToken, authLimiter: newAuthLimiter()}
}

func (h *FileUploadHandler) Handle(w http.ResponseWriter, r *http.Request) {
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
	if err := r.ParseMultipartForm(32 << 20); err != nil {
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
	id := uuid.New().String()
	storeDir := filepath.Join(h.UploadDir, id)
	if err := fileioutil.CreateDirIfNotExist(storeDir, 0750); err != nil {
		slog.Error("Unable to create upload dir", "error", err)
		internalServerError(w)
		return
	}
	storePath := filepath.Join(storeDir, filename)
	if err := streamToFile(file, storePath); err != nil {
		slog.Error("Unable to store file", "error", err)
		internalServerError(w)
		return
	}
	slog.Info("Stored upload", "storePath", storePath)
	okWithJson(w, &FileUploadResponse{Id: id})
}
