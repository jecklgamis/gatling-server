package uploader

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func UploadFile(uploadUrl string, filename string, kv map[string]string) (*http.Response, error) {
	req, err := CreateMultipartRequest(uploadUrl, filename, kv)
	if err != nil {
		return nil, err
	}
	hc := &http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	log.Printf("Uploaded %s to %s\n", filename, uploadUrl)
	return resp, nil
}

func CreateMultipartRequest(uploadURL string, filename string, kv map[string]string) (*http.Request, error) {
	var body = &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if filename != "" {
		file, err := os.Open(filename)
		if err != nil {
			log.Println("Unable to open file", filename)
			return nil, err
		}
		defer file.Close()
		part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
		if err != nil {
			log.Println("Unable to create form", err)
			return nil, err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			log.Println("Unable to copy file", err)
			return nil, err
		}
	}
	for k, v := range kv {
		part, err := writer.CreateFormField(k)
		if err != nil {
			log.Println("Unable to create form field", k, err)
			return nil, err
		}
		_, err = io.Copy(part, strings.NewReader(v))
		if err != nil {
			log.Println("Unable to set field", k, err)
			return nil, err
		}
	}
	writer.Close()
	request, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		log.Println("Unable to create request", err)
		return nil, err
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	return request, nil
}
