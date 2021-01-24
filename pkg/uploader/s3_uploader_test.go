package uploader

import (
	"fmt"
	"github.com/jecklgamis/gatling-server/pkg/s3"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"github.com/jecklgamis/gatling-server/pkg/workspace"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestCreateUploaderWithEmptyConfig(t *testing.T) {
	uploader := NewS3Uploader(s3.NewS3Manager("some-region"), map[string]string{})
	test.Assert(t, uploader == nil, "expecting nil uploader")
}

func TestUploadSuccessful(t *testing.T) {
	uploader := NewS3Uploader(s3.NewFakeS3Ops([]byte("some-data"), "some-path", nil),
		map[string]string{"s3url": "s3://some-bucket/some-key"})
	test.Assert(t, uploader != nil, "expecting valid uploader")
	err := uploader.Upload("someTaskId", someUserFilesDir())
	test.Assert(t, err == nil, "expecting upload error")
}

func TestUploadFailure(t *testing.T) {
	uploader := NewS3Uploader(s3.NewFakeS3Ops([]byte("some-data"), "some-path", fmt.Errorf("some-error")),
		map[string]string{"s3url": "s3://some-bucket"})
	test.Assert(t, uploader != nil, "expecting valid uploader")
	err := uploader.Upload("someTaskId", someUserFilesDir())
	test.Assert(t, err != nil, "expecting upload error")
}

func someUserFilesDir() *workspace.UserFilesDir {
	dir, _ := ioutil.TempDir("", "")
	userFilesDir, _ := workspace.NewUserFilesDir(filepath.Join(dir, "user-files-dir"))
	return userFilesDir
}
