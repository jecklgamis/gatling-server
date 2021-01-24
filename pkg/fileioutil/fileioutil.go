package fileioutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func MustReadFile(filename string) []byte {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return content
}

func WriteBufferToFile(buffer *bytes.Buffer, dir string, filename string) (*string, error) {
	if !DirExists(dir) {
		return nil, fmt.Errorf("dir %v does not exist", dir)
	}
	err := os.MkdirAll(dir, 0744)
	if err != nil {
		return nil, err
	}
	storePath := filepath.Join(dir, filename)
	err = ioutil.WriteFile(storePath, buffer.Bytes(), 0744)
	if err != nil {
		return nil, err
	}
	log.Println("Wrote", storePath)
	return &storePath, nil
}

func CreateDirIfNotExist(path string, perm os.FileMode) error {
	if DirExists(path) {
		return nil
	}
	if err := os.MkdirAll(path, perm); err != nil {
		return nil
	}
	log.Println("Created", path)
	return nil
}

func DirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return fi.IsDir()
}

func FileExist(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return !fi.IsDir()
}

func CopyFile(src string, dst string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(dst, input, 0744)
	if err != nil {
		return err
	}
	return nil
}

func FindFile(dir string, filename string) (foundPath string, err error) {
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.Name() == filename {
			foundPath = path
			return nil
		}
		return nil
	})
	if foundPath == "" {
		return "", fmt.Errorf("file not found")
	}
	return foundPath, nil
}
