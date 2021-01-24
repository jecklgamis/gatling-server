package workspace

import (
	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateWorkspace(t *testing.T) {
	createWorkSpace(t)
}

func TestCreateUserFilesDir(t *testing.T) {
	workspace := createWorkSpace(t)
	defer func() {
		os.RemoveAll(workspace.baseDir)
	}()
	userFilesDir, err := workspace.NewUserFilesDir("some-user-files-dir")
	test.Assert(t, err == nil, "unable to user files dir")
	test.Assertf(t, filepath.Dir(userFilesDir.BaseDir) == workspace.BaseDir(), "unexpected parent dir")
}

func TestReadFile(t *testing.T) {
	workspace := createWorkSpace(t)
	defer func() {
		os.RemoveAll(workspace.baseDir)
	}()
	userFilesDir, err := workspace.NewUserFilesDir("some-id")
	test.Assert(t, err == nil, "unable to create user files dir")
	consoleLog := filepath.Join(userFilesDir.BaseDir, "console.log")
	logContent := "some log"

	err = ioutil.WriteFile(consoleLog, []byte(logContent), 0744)
	test.Assert(t, err == nil, "unable to create file")
	test.Assert(t, fileioutil.FileExist(consoleLog), "expecting console.log to exist")
	content, err := workspace.ReadFile("some-id", "console.log")
	test.Assert(t, string(content) == logContent, "unable read console.log")
}

func createWorkSpace(t *testing.T) *Workspace {
	dir, err := ioutil.TempDir("", "workspace")
	test.Assert(t, err == nil, "unable to create tmp dir")

	workspace := NewWorkspace(dir)
	test.Assert(t, workspace != nil, "unable to create workspace")

	return workspace
}
