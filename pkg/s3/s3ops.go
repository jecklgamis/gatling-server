package s3

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
)

type S3Manager struct {
	region string
}

func NewS3Manager(region string) *S3Manager {
	return &S3Manager{region: region}
}

type S3UploadOps interface {
	Upload(bucket string, key string, filename string) error
}

type S3DownloadOps interface {
	Download(bucket string, key string, dstDir string) (*string, error)
	DownloadUrl(url string, dstDir string) (*string, error)
}

type S3Ops interface {
	S3UploadOps
	S3DownloadOps
}

func (s *S3Manager) Upload(bucket string, key string, filename string) error {
	log.Printf("Uploading %s to %s/%s", filename, bucket, key)
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	s3Client := s3.New(sess)
	uploader := s3manager.NewUploaderWithClient(s3Client)
	file, err := os.Open(filename)
	if err != nil {
		log.Println("Unable to open file :", err)
		return err
	}
	uploadParams := &s3manager.UploadInput{Bucket: &bucket, Key: &key, Body: file}
	result, err := uploader.Upload(uploadParams)
	if err != nil {
		log.Println("Unable to upload file :", err)
		return err
	}
	log.Println("Uploaded to", result.Location)
	return nil
}

func (s *S3Manager) Download(bucket string, key string, dir string) (*string, error) {
	if !fileioutil.DirExists(dir) {
		return nil, fmt.Errorf("destination dir does not exist")
	}
	log.Printf("Downloading %s from %s", key, bucket)
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(s.region)},
		SharedConfigState: session.SharedConfigEnable,
	}))
	downloader := s3manager.NewDownloader(sess)
	filename := filepath.Base(key)
	storePath := filepath.Join(dir, filename)
	file, err := os.Create(storePath)
	if err != nil {
		log.Println("Failed to create file : ", err)
		return nil, err
	}
	defer file.Close()
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
	if err != nil {
		return nil, err
	}
	log.Printf("Downloaded to %s (%d bytes)", storePath, numBytes)
	return &storePath, nil
}

func (s *S3Manager) DownloadUrl(s3url, dir string) (*string, error) {
	bucket, key, err := ParseS3Uri(s3url)
	if err != nil {
		return nil, err
	}
	return s.Download(bucket, key, dir)
}

func ParseS3Uri(uri string) (bucket string, key string, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", "", err
	}
	var path string
	if len(u.Path) > 0 {
		path = u.Path[1:]
	}
	return u.Host, path, nil
}
