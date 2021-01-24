package taskmanager

import (
	"fmt"
	"github.com/jecklgamis/gatling-server/pkg/jsonutil"
	"io/ioutil"
	"log"
	"path/filepath"
)

type TaskStoreOps interface {
	Store(context *TaskRuntimeContext) error
	Load(id string) (*TaskRuntimeContext, bool)
}

type FileTaskStore struct {
	baseDir string
}

func NewFileTaskStore(baseDir string) *FileTaskStore {
	return &FileTaskStore{baseDir: baseDir}
}

func (s *FileTaskStore) Store(context *TaskRuntimeContext) error {
	path := filepath.Join(s.baseDir, fmt.Sprintf("%s-context.json", context.Task.Id))
	err := ioutil.WriteFile(path, []byte(jsonutil.ToJson(context)), 0744)
	log.Println("Wrote", path)
	if err != nil {
		return err
	}
	return nil
}

func (s *FileTaskStore) Load(id string) (*TaskRuntimeContext, bool) {
	return nil, false
}
