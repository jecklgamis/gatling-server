package s3

import (
	"io/ioutil"
	"log"
)

type FakeS3Ops struct {
	downloadData []byte
	err          error
	storePath    string
}

func NewFakeS3Ops(downloadData []byte, storePath string, err error) *FakeS3Ops {
	return &FakeS3Ops{downloadData: downloadData, storePath: storePath, err: err}
}

func (f *FakeS3Ops) Upload(_ string, _ string, _ string) error {
	return f.err
}

func (f *FakeS3Ops) DownloadUrl(string, string) (*string, error) {
	err := ioutil.WriteFile(f.storePath, f.downloadData, 0744)
	if err != nil {
		log.Println(err)
	}
	return &f.storePath, nil
}

func (f *FakeS3Ops) Download(string, string, string) (*string, error) {
	err := ioutil.WriteFile(f.storePath, f.downloadData, 0744)
	if err != nil {
		log.Println(err)
	}
	return &f.storePath, nil
}
