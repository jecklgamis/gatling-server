package taskmanager

import (
	"fmt"
	"github.com/jecklgamis/gatling-server/pkg/cmdexec"
	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	"github.com/jecklgamis/gatling-server/pkg/gatling"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"github.com/jecklgamis/gatling-server/pkg/uploader"
	"github.com/jecklgamis/gatling-server/pkg/workspace"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestSubmitTask(t *testing.T) {
	task := createSomeGatlingTask()
	taskManager := NewTaskManager(fakeGatlingOps(0), make(chan interface{}, 1024), []uploader.GatlingArtifactUploader{})
	resultC, err := taskManager.SubmitTask(task)
	test.Assertf(t, err == nil, "failed to submit task : %v", err)
	result := <-resultC
	test.Assertf(t, result.Ok, "expecting task to succeed")
	taskContext, found := taskManager.GetTaskRuntimeContext(task.Id)
	test.Assertf(t, found, "task context not fund for task %v", task.Id)
	test.Assertf(t, taskContext != nil, "unable get task context for task %v", task.Id)
	test.Assertf(t, taskContext.Success, "expecting successful result")
	test.Assertf(t, taskContext.Status == TaskCompleted, "unexpected status %v", taskContext.Status)
	test.Assertf(t, taskContext.Started.String() != "0001-01-01 00:00:00 +0000 UTC", "start time not set")
	test.Assertf(t, taskContext.Completed.String() != "0001-01-01 00:00:00 +0000 UTC", "completion time not set")
}

func TestAbortTaskAfterCompletionShouldFail(t *testing.T) {
	task := createSomeGatlingTask()
	taskManager := NewTaskManager(fakeGatlingOps(0), make(chan interface{}, 1024), []uploader.GatlingArtifactUploader{})
	resultC, err := taskManager.SubmitTask(task)
	test.Assertf(t, err == nil, "failed to submit task : %v", err)
	result := <-resultC
	test.Assertf(t, result.Ok, "expecting task to succeed")
	err = taskManager.AbortTask(task.Id)
	test.Assertf(t, err != nil, "expecting to fail")
	taskContext, found := taskManager.GetTaskRuntimeContext(task.Id)
	test.Assertf(t, found, "task context not found for task %v", task.Id)
	test.Assertf(t, taskContext.Status == TaskCompleted, "unexpected status %v", taskContext.Status)
}

func TestAbortTask(t *testing.T) {
	task := createSomeGatlingTask()
	tm := NewTaskManager(fakeGatlingOps(5*time.Second), make(chan interface{}, 1024), []uploader.GatlingArtifactUploader{})

	_, err := tm.SubmitTask(task)
	test.Assertf(t, err == nil, "failed to submit task : %v", err)
	go func() {
		time.Sleep(2 * time.Second)
		err := tm.AbortTask(task.Id)
		test.Assertf(t, err == nil, "expecting abort to succeed but got %v", err)
	}()
	taskContext, found := tm.GetTaskRuntimeContext(task.Id)
	test.Assertf(t, found, "task context not found for task %v", task.Id)
	time.Sleep(5 * time.Second)
	test.Assertf(t, taskContext.Status == TaskAborted, "unexpected status %v", taskContext.Status)
}

func createSomeGatlingTask() *gatling.Task {
	tmpDir, _ := ioutil.TempDir("", "")
	userFilesDir, _ := workspace.NewUserFilesDir(filepath.Join(tmpDir, "repos"))
	fileioutil.CopyFile("testdata/SingleFileExampleSimulation.scala",
		fmt.Sprintf("%s/SingleFileExampleSimulation.scala",
			userFilesDir.Simulations))
	return gatling.NewTask(CreateTaskId(), "gatling.test.example.simulation.SingleFileExampleSimulation",
		"-DbaseUrl=http://localhost:8080 -DdurationMin=0.10 -DrequestPerSecond=1", userFilesDir)
}

func fakeGatlingOps(delay time.Duration) gatling.Ops {
	return gatling.RunSimulationFunc(func(commandOps cmdexec.CommandExecutionOps, task *gatling.Task) (*exec.Cmd, error) {
		cmd := exec.Command("sleep", fmt.Sprintf("%v", delay))
		err := commandOps.Execute(cmd)
		if err != nil {
			return nil, err
		}
		return cmd, nil
	})
}
