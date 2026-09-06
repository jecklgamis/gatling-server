package handler

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	"log"
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
		log.Println("Upload dir is not absolute")
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
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
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
	id := uuid.New().String()
	storeDir := filepath.Join(h.UploadDir, id)
	if err := fileioutil.CreateDirIfNotExist(storeDir, 0750); err != nil {
		log.Println("Unable to create upload dir :", err)
		internalServerError(w)
		return
	}
	storePath := filepath.Join(storeDir, filename)
	if err := streamToFile(file, storePath); err != nil {
		log.Println("Unable to store file :", err)
		internalServerError(w)
		return
	}
	log.Println("Stored upload", storePath)
	okWithJson(w, &FileUploadResponse{Id: id})
}
