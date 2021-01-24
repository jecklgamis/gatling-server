package uploader

import (
	"github.com/jecklgamis/gatling-server/pkg/s3"
	"github.com/jecklgamis/gatling-server/pkg/workspace"
	"log"
	"path/filepath"
)

type S3Uploader struct {
	configMap map[string]string
	s3Ops     s3.S3Ops
}

func NewS3Uploader(s3Ops s3.S3Ops, configMap map[string]string) *S3Uploader {
	if _, ok := configMap["s3url"]; !ok {
		log.Println("no s3url found in config map")
		return nil
	}
	return &S3Uploader{configMap: configMap, s3Ops: s3Ops}
}

func (u *S3Uploader) Upload(taskId string, userFilesDir *workspace.UserFilesDir) error {
	bucket, key, err := s3.ParseS3Uri(u.configMap["s3url"])
	if err != nil {
		log.Println("Failed to upload artifacts :", err)
		return err
	}
	err = u.s3Ops.Upload(bucket, filepath.Join(key, taskId), filepath.Join(userFilesDir.BaseDir, "results.tar.gz"))
	if err != nil {
		log.Println("Failed to upload results :", err)
		return err
	}
	return nil
}
