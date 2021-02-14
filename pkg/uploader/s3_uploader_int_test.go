package uploader

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/jecklgamis/gatling-server/pkg/env"
	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	"github.com/jecklgamis/gatling-server/pkg/s3"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestUploadResults(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	region := env.GetOrPanic("AWS_REGION")
	s3Url := env.GetOrPanic("GATLING_SERVER_RESULTS_S3_URL")
	s3Ops := s3.NewS3Manager(region)
	configMap := map[string]string{"s3url": s3Url}
	uploader := NewS3Uploader(s3Ops, configMap)

	userFilesDir := someUserFilesDir()
	resultsTarGzPath := filepath.Join(userFilesDir.BaseDir, "results.tar.gz")
	println("resultsTarGzPath", resultsTarGzPath)
	err := ioutil.WriteFile(resultsTarGzPath, []byte("some-data"), 0744)
	test.Assertf(t, err == nil, "failed to create test data :%v", err)

	someTaskId := uuid.New().String()[0:8]
	err = uploader.Upload(someTaskId, userFilesDir)
	test.Assertf(t, err == nil, "failed to upload : %v", err)

	//verify results.tar.gz exist in s3
	downloadPath, err := s3Ops.DownloadUrl(fmt.Sprintf("%s/%s/results.tar.gz", s3Url, someTaskId),
		userFilesDir.BaseDir)
	test.Assertf(t, err == nil, "failed to download results.tar.gz : %v", err)

	test.Assertf(t, fileioutil.FileExist(*downloadPath), "failed to verify uploaded results.tar.gz : %v", err)
}
