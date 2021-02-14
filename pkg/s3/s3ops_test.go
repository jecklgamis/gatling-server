package s3

import (
	"github.com/jecklgamis/gatling-server/pkg/env"
	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"io/ioutil"
	"testing"
)

func TestUploadS3File(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	s3Ops := NewS3Manager(env.GetOrPanic("AWS_REGION"))
	bucket, _, err := ParseS3Uri(env.GetOrPanic("GATLING_SERVER_INCOMING_S3_URL"))
	err = s3Ops.Upload(bucket,
		"some-task-id/some.txt", "testdata/some.txt")
	test.Assertf(t, err == nil, "failed to upload file to s3")
}

func TestDownloadS3File(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dir, _ := ioutil.TempDir("", "")
	s3Ops := NewS3Manager(env.GetOrPanic("AWS_REGION"))
	bucket, _, err := ParseS3Uri(env.GetOrPanic("GATLING_SERVER_INCOMING_S3_URL"))
	storePath, err := s3Ops.Download(bucket, "some-folder/SingleFileExampleSimulation.scala", dir)
	test.Assertf(t, err == nil, "failed to download file to s3 : %v", err)
	test.Assertf(t, fileioutil.FileExist(*storePath), "expecting a file to exist")
}

func TestParseS3URI(t *testing.T) {
	uri := "s3://some-bucket/some-folder/SingleFileExampleSimulation.scala"
	bucket, key, err := ParseS3Uri(uri)
	test.Assertf(t, err == nil, "failed to parse s3 uri : %v", err)
	test.Assertf(t, bucket == "some-bucket", "unexpected bucket %v", bucket)
	test.Assertf(t, key == "some-folder/SingleFileExampleSimulation.scala", "unexpected key %v", key)
}
