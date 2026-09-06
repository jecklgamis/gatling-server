package api

// SubmitTaskRequest is a generic task submission request. Url may point to
// either an http(s) location or an s3:// location; the handler dispatches
// the download based on the URL scheme.
type SubmitTaskRequest struct {
	Simulation string `json:"simulation"`
	JavaOpts   string `json:"javaOpts"`
	Url        string `json:"url"`
}

type S3DownloadTaskRequest struct {
	Simulation string `json:"simulation"`
	JavaOpts   string `json:"javaOpts"`
	Url        string `json:"url"`
}

type FileUploadTaskRequest struct {
	Simulation string `json:"simulation"`
	JavaOpts   string `json:"javaOpts"`
	Url        string `json:"url"`
}

type SubmitTaskResponse struct {
	Ok     bool   `json:"ok"`
	TaskId string `json:"taskId"`
}
