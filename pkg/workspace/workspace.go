package workspace

import (
	server "github.com/jecklgamis/gatling-server/pkg/fileioutil"
	"io/ioutil"
	"log"
	"path/filepath"
)

type Ops interface {
	NewUserFilesDir(id string) (*UserFilesDir, error)
	BaseDir() string
	ReadFile(id, filename string) ([]byte, error)
}

type Workspace struct {
	baseDir string
}

func NewWorkspace(baseDir string) *Workspace {
	if err := server.CreateDirIfNotExist(baseDir, 0744); err != nil {
		panic(err)
	}
	log.Println("Created workspace", baseDir)
	return &Workspace{baseDir: baseDir}
}

func (w *Workspace) NewUserFilesDir(id string) (*UserFilesDir, error) {
	baseDir := filepath.Join(w.baseDir, id)
	return NewUserFilesDir(baseDir)
}

func (w *Workspace) ReadFile(id, filename string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(w.baseDir, id, filename))
}

func (r *Workspace) BaseDir() string {
	return r.baseDir
}
