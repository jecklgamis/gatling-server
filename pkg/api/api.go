package api

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
