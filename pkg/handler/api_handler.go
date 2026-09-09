package handler

import (
	"encoding/json"
	"fmt"
	"github.com/jecklgamis/gatling-server/pkg/api"
	"github.com/jecklgamis/gatling-server/pkg/gatling"
	"github.com/jecklgamis/gatling-server/pkg/s3"
	"github.com/jecklgamis/gatling-server/pkg/taskmanager"
	"github.com/jecklgamis/gatling-server/pkg/workspace"
	"io"

	"log/slog"
	"net"
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

// HttpHostAuth is the credential to attach when downloading from a specific
// allowlisted host. Type is "basic" (Username/Password) or "bearer" (Token).
type HttpHostAuth struct {
	Type     string
	Username string
	Password string
	Token    string
}

// AllowedHttpHost is one entry in ApiHandler's http(s) download allowlist.
// Auth is optional; when nil, a host still gets the handler's own
// BrowseUsername/BrowsePassword attached (see ApiHandler doc comment) rather
// than nothing, so the self-referential /uploads flow keeps working without
// needing explicit per-host config.
type AllowedHttpHost struct {
	Host string
	Auth *HttpHostAuth
}

// DefaultAllowedHttpHosts is used when no explicit allowlist is configured,
// so the documented "upload then submit against http://<self>/uploads/..."
// flow keeps working out of the box.
var DefaultAllowedHttpHosts = []AllowedHttpHost{{Host: "localhost"}, {Host: "127.0.0.1"}, {Host: "::1"}}

// ApiHandler accepts a generic task submission request whose Url may point to
// either an http(s) location or an s3:// location, downloads the referenced
// jar accordingly, and submits it as a task.
//
// Both download paths are scoped to prevent the server from being used as an
// SSRF pivot with a valid API token: http(s) downloads are restricted to
// AllowedHttpHosts (anything else is only allowed if it resolves to a public,
// non-private/link-local/loopback address), and s3 downloads are restricted
// to AllowedS3Buckets.
type ApiHandler struct {
	WorkspaceOps     workspace.Ops
	TaskOps          taskmanager.Ops
	s3Ops            s3.S3Ops
	ApiToken         string
	AllowedHttpHosts []AllowedHttpHost
	AllowedS3Buckets []string
	// BrowseUsername/BrowsePassword are the HTTP Basic Auth credentials
	// gating /uploads/, attached when downloading from an AllowedHttpHosts
	// entry that doesn't specify its own Auth (the self-referential "upload
	// then submit against http://<self>/uploads/..." flow) so that flow
	// keeps working now that /uploads/ requires auth. Never sent to hosts
	// outside the explicit allowlist, so this can't leak to a third party.
	BrowseUsername string
	BrowsePassword string
	authLimiter    *authLimiter
}

func NewApiHandler(workspaceOps workspace.Ops, taskOps taskmanager.Ops, s3Ops s3.S3Ops, apiToken string,
	allowedHttpHosts []AllowedHttpHost, allowedS3Buckets []string, browseUsername string, browsePassword string) *ApiHandler {
	if allowedHttpHosts == nil {
		allowedHttpHosts = DefaultAllowedHttpHosts
	}
	return &ApiHandler{WorkspaceOps: workspaceOps, TaskOps: taskOps, s3Ops: s3Ops, ApiToken: apiToken,
		AllowedHttpHosts: allowedHttpHosts, AllowedS3Buckets: allowedS3Buckets,
		BrowseUsername: browseUsername, BrowsePassword: browsePassword, authLimiter: newAuthLimiter()}
}

func (h *ApiHandler) Handle(w http.ResponseWriter, r *http.Request) {
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
	if r.Body == nil {
		badRequestWithError(w, fmt.Errorf("body is nil"))
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxApiTaskRequestSize+1))
	if err != nil {
		slog.Error("Unable to read request body", "error", err)
		internalServerError(w)
		return
	}
	if len(body) > maxApiTaskRequestSize {
		badRequestWithError(w, fmt.Errorf("request body too large"))
		return
	}
	request := api.SubmitTaskRequest{}
	if err := json.Unmarshal(body, &request); err != nil {
		slog.Error("Unable to marshall request body", "error", err)
		badRequestWithError(w, fmt.Errorf("unable to marshall request body"))
		return
	}
	if err := validateSubmitTaskRequest(&request); err != nil {
		slog.Error("Invalid request", "error", err)
		badRequestWithError(w, err)
		return
	}
	taskId := taskmanager.CreateTaskId()
	taskPath := filepath.Join(h.WorkspaceOps.BaseDir(), taskId)
	userFilesDir, err := workspace.NewUserFilesDir(taskPath)
	if err != nil {
		slog.Error("Unable to create user files directory", "error", err)
		internalServerError(w)
		return
	}
	taskCommitted := false
	defer func() {
		if !taskCommitted {
			if err := os.RemoveAll(taskPath); err != nil {
				slog.Error("Unable to remove task dir after failed download", "error", err)
			}
		}
	}()

	storePath, err := h.download(request.Url, userFilesDir.Simulations)
	if err != nil {
		slog.Error("Unable to download file", "error", err)
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
	slog.Info("Submitting simulation", "filename", filename)
	metadata := &Metadata{TaskId: taskId, Simulation: request.Simulation, JavaOpts: request.JavaOpts}
	if err := writeMetadata(userFilesDir.BaseDir, metadata, "metadata.json"); err != nil {
		slog.Error("Unable write metadata file", "error", err)
		internalServerError(w)
		return
	}
	if _, err := h.TaskOps.SubmitTask(task); err != nil {
		slog.Error("Unable to submit task", "error", err)
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
		bucket, _, err := s3.ParseS3Uri(rawUrl)
		if err != nil {
			return nil, err
		}
		if !isAllowedBucket(bucket, h.AllowedS3Buckets) {
			return nil, fmt.Errorf("bucket %q is not in the allowed s3 bucket list", bucket)
		}
		return h.s3Ops.DownloadUrl(rawUrl, dstDir)
	case "http", "https":
		entry, err := checkHttpHostAllowed(u.Hostname(), h.AllowedHttpHosts)
		if err != nil {
			return nil, err
		}
		var auth *HttpHostAuth
		if entry != nil {
			if entry.Auth != nil {
				auth = entry.Auth
			} else if h.BrowseUsername != "" || h.BrowsePassword != "" {
				auth = &HttpHostAuth{Type: "basic", Username: h.BrowseUsername, Password: h.BrowsePassword}
			}
		}
		return downloadHttpFile(rawUrl, dstDir, auth)
	default:
		return nil, fmt.Errorf("unsupported url scheme %q", u.Scheme)
	}
}

// isAllowedBucket reports whether bucket is in allowedBuckets. An empty
// allowedBuckets list denies every bucket - the s3 downloader must be
// explicitly scoped before it can be used.
func isAllowedBucket(bucket string, allowedBuckets []string) bool {
	for _, b := range allowedBuckets {
		if strings.EqualFold(b, bucket) {
			return true
		}
	}
	return false
}

// checkHttpHostAllowed rejects hosts that would let an authenticated caller
// pivot the server into fetching internal/cloud-metadata resources. A host
// explicitly present in allowedHosts is always permitted (this is how the
// documented "submit a URL pointing back at my own /uploads" flow keeps
// working) and its allowlist entry (including any per-host Auth) is
// returned; anything else must resolve only to public IP addresses, and gets
// no entry (and therefore no credentials) back.
func checkHttpHostAllowed(host string, allowedHosts []AllowedHttpHost) (*AllowedHttpHost, error) {
	for i := range allowedHosts {
		if strings.EqualFold(allowedHosts[i].Host, host) {
			return &allowedHosts[i], nil
		}
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve host %q: %w", host, err)
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return nil, fmt.Errorf("host %q resolves to a disallowed address %s", host, ip)
		}
	}
	return nil, nil
}

// downloadHttpFile fetches rawUrl into dstDir. When auth is non-nil, its
// credentials are attached to the outgoing request - the caller is
// responsible for only passing auth belonging to a trusted, explicitly
// allow-listed host (see checkHttpHostAllowed).
func downloadHttpFile(rawUrl string, dstDir string, auth *HttpHostAuth) (*string, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	filename := filepath.Base(u.Path)
	if filename == "" || filename == "." || filename == "/" {
		return nil, fmt.Errorf("unable to determine filename from url")
	}
	req, err := http.NewRequest(http.MethodGet, rawUrl, nil)
	if err != nil {
		return nil, err
	}
	if auth != nil {
		switch strings.ToLower(auth.Type) {
		case "basic":
			req.SetBasicAuth(auth.Username, auth.Password)
		case "bearer":
			req.Header.Set("Authorization", "Bearer "+auth.Token)
		default:
			slog.Warn("Unrecognized http host auth type, downloading unauthenticated", "type", auth.Type)
		}
	}
	client := &http.Client{Timeout: httpDownloadTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("Unable to close response body", "error", err)
		}
	}()
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
		if err := os.Remove(storePath); err != nil {
			slog.Error("Unable to remove oversized download", "error", err)
		}
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
