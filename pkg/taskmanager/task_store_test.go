package taskmanager

import (
	"fmt"
	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	"github.com/jecklgamis/gatling-server/pkg/gatling"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"github.com/jecklgamis/gatling-server/pkg/workspace"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestStoreAndLoad(t *testing.T) {
	dir, _ := ioutil.TempDir("", "task-store")
	store := NewFileTaskStore(dir)
	err := store.Store(someTaskContext(t))
	test.Assertf(t, err == nil, "failed to store task context")
}

func someTaskContext(t *testing.T) *TaskRuntimeContext {
	return &TaskRuntimeContext{Task: someGatlingTask(t, "http://localhost:8080")}
}

func someGatlingTask(t *testing.T, targetUrl string) *gatling.Task {
	userFilesPath, _ := ioutil.TempDir("", "")
	userFilesDir, _ := workspace.NewUserFilesDir(filepath.Join(userFilesPath, "user-files-dir"))
	test.Assertf(t, userFilesDir != nil, "nil user files dir")
	fileioutil.CopyFile("testdata/SingleFileExampleSimulation.scala",
		fmt.Sprintf("%s/SingleFileExampleSimulation.scala",
			userFilesDir.Simulations))
	return gatling.NewTask("some-task-id", "gatling.test.example.simulation.SingleFileExampleSimulation",
		fmt.Sprintf("-DbaseUrl=%s -DdurationMin=0.10 -DrequestPerSecond=1", targetUrl), userFilesDir)
}
