package uploader

import (
	"bytes"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func UploadFile(uploadUrl string, filename string, kv map[string]string, headers map[string]string) (*http.Response, error) {
	req, err := CreateMultipartRequest(uploadUrl, filename, kv, headers)
	if err != nil {
		return nil, err
	}
	hc := &http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	slog.Info("Uploaded", "filename", filename, "uploadUrl", uploadUrl)
	return resp, nil
}

func CreateMultipartRequest(uploadURL string, filename string, kv map[string]string, headers map[string]string) (*http.Request, error) {
	var body = &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if filename != "" {
		file, err := os.Open(filename)
		if err != nil {
			slog.Error("Unable to open file", "filename", filename, "error", err)
			return nil, err
		}
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				slog.Error("Unable to close file", "error", closeErr)
			}
		}()
		part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
		if err != nil {
			slog.Error("Unable to create form", "error", err)
			return nil, err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			slog.Error("Unable to copy file", "error", err)
			return nil, err
		}
	}
	for k, v := range kv {
		part, err := writer.CreateFormField(k)
		if err != nil {
			slog.Error("Unable to create form field", "field", k, "error", err)
			return nil, err
		}
		_, err = io.Copy(part, strings.NewReader(v))
		if err != nil {
			slog.Error("Unable to set field", "field", k, "error", err)
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		slog.Error("Unable to close multipart writer", "error", err)
		return nil, err
	}
	request, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		slog.Error("Unable to create request", "error", err)
		return nil, err
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	for k, v := range headers {
		request.Header.Add(k, v)
	}
	return request, nil
}
