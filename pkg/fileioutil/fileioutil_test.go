package fileioutil

import (
	"bytes"
	"github.com/google/uuid"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	dir, _ := ioutil.TempDir("", "")
	src := "testdata/gatling-test-example-user-files.tar.gz"
	dst := filepath.Join(dir, "gatling-test-example-user-files.tar.gz")
	err := CopyFile(src, dst)
	test.Assertf(t, err == nil, "failed to copy file : %v", err)
	test.Assertf(t, FileExist(dst), "expected file to exist")
}

func TestCreateDirIfNotExist(t *testing.T) {
	temp, _ := ioutil.TempDir("", "")
	dst := filepath.Join(temp, "some-temp")
	test.Assertf(t, !DirExists(dst), "expected non existent dir")
	test.Assertf(t, CreateDirIfNotExist(dst, 0744) == nil, "unable to create dir")
	test.Assertf(t, DirExists(dst), "expected dir to exist")
}

func TestWriteBufferToFile(t *testing.T) {
	buffer := bytes.NewBuffer([]byte("some data"))
	dir, _ := ioutil.TempDir("", "")
	path, _ := WriteBufferToFile(buffer, dir, "temp.txt")
	data, _ := ioutil.ReadFile(*path)
	test.Assertf(t, string(data) == "some data", "unexpected data %v", string(data))
}

func TestFileExist(t *testing.T) {
	file, _ := ioutil.TempFile("", "")
	test.Assertf(t, FileExist(file.Name()), "expecting file to exist")
	dir, _ := ioutil.TempDir("", "")
	test.Assertf(t, !FileExist(filepath.Join(dir, "some.txt")), "not expecting file to exist")
}

func TestWriteBufferToFileFailsOnDir(t *testing.T) {
	dir, _ := ioutil.TempDir("", "")
	test.Assertf(t, !FileExist(dir), "expecting dir to fail")
}

func TestWriteBufferToFileFailsOnNonExistentDir(t *testing.T) {
	dir, _ := ioutil.TempDir("", "")
	_, err := WriteBufferToFile(&bytes.Buffer{}, filepath.Join(dir, "some-path"), "some-file.txt")
	test.Assertf(t, err != nil, "expecting to fail")
}

func TestFindFile(t *testing.T) {
	dir, _ := ioutil.TempDir("", "")
	subDir := filepath.Join(dir, "path1", "path2")
	err := CreateDirIfNotExist(subDir, 0744)
	test.Assertf(t, err == nil, "unable to create dirs test file :%v", err)

	err = ioutil.WriteFile(filepath.Join(subDir, "some.txt"), []byte("some-text"), 0744)
	test.Assertf(t, err == nil, "unable to create test file :%v", err)

	path, err := FindFile(dir, "some.txt")
	test.Assertf(t, err == nil, "%v", err)
	test.Assertf(t, filepath.Join(subDir, "some.txt") == path, "invalid path returned")
}

func TestMustReadFile(t *testing.T) {
	var panicCaught = false
	defer func() {
		if r := recover(); r != nil {
			log.Println("panic!")
			panicCaught = true
		}
		test.Assertf(t, panicCaught, "expecting panic")
	}()
	MustReadFile(uuid.New().String())
}
