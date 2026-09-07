package taskmanager

import (
	"fmt"
	"github.com/jecklgamis/gatling-server/pkg/cmdexec"
	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	"github.com/jecklgamis/gatling-server/pkg/gatling"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"github.com/jecklgamis/gatling-server/pkg/uploader"
	"github.com/jecklgamis/gatling-server/pkg/workspace"
	"os"

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

func TestSubmitNilTaskReturnsError(t *testing.T) {
	tm := NewTaskManager(fakeGatlingOps(0), make(chan interface{}, 1024), []uploader.GatlingArtifactUploader{})
	_, err := tm.SubmitTask(nil)
	test.Assertf(t, err != nil, "expecting error for nil task")
}

func TestAbortUnknownTaskReturnsError(t *testing.T) {
	tm := NewTaskManager(fakeGatlingOps(0), make(chan interface{}, 1024), []uploader.GatlingArtifactUploader{})
	err := tm.AbortTask("some-unknown-task-id")
	test.Assertf(t, err != nil, "expecting error for unknown task")
}

func TestGetTaskRuntimeContextUnknownTaskReturnsFalse(t *testing.T) {
	tm := NewTaskManager(fakeGatlingOps(0), make(chan interface{}, 1024), []uploader.GatlingArtifactUploader{})
	context, found := tm.GetTaskRuntimeContext("some-unknown-task-id")
	test.Assertf(t, !found, "expecting task not found")
	test.Assertf(t, context == nil, "expecting nil context")
}

func TestAbortTaskWithoutProcessReturnsError(t *testing.T) {
	tm := NewTaskManager(fakeGatlingOps(0), make(chan interface{}, 1024), []uploader.GatlingArtifactUploader{})
	task := createSomeGatlingTask()
	tm.storeContext(&TaskRuntimeContext{Task: task, Status: TaskStarted})
	err := tm.AbortTask(task.Id)
	test.Assertf(t, err != nil, "expecting error since no process is set")
}

func TestTaskFailsWhenRunSimulationReturnsError(t *testing.T) {
	task := createSomeGatlingTask()
	tm := NewTaskManager(fakeGatlingOpsWithError(fmt.Errorf("some-error")), make(chan interface{}, 1024),
		[]uploader.GatlingArtifactUploader{})
	resultC, err := tm.SubmitTask(task)
	test.Assertf(t, err == nil, "failed to submit task : %v", err)
	result := <-resultC
	test.Assertf(t, !result.Ok, "expecting task to fail")
	taskContext, found := tm.GetTaskRuntimeContext(task.Id)
	test.Assertf(t, found, "task context not found for task %v", task.Id)
	test.Assertf(t, taskContext.Status == TaskCompleted, "unexpected status %v", taskContext.Status)
	test.Assertf(t, !taskContext.Success, "expecting unsuccessful result")
}

func TestTaskFailsWhenProcessExitsNonZero(t *testing.T) {
	task := createSomeGatlingTask()
	tm := NewTaskManager(fakeFailingGatlingOps(), make(chan interface{}, 1024), []uploader.GatlingArtifactUploader{})
	resultC, err := tm.SubmitTask(task)
	test.Assertf(t, err == nil, "failed to submit task : %v", err)
	result := <-resultC
	test.Assertf(t, !result.Ok, "expecting task to fail")
	taskContext, found := tm.GetTaskRuntimeContext(task.Id)
	test.Assertf(t, found, "task context not found for task %v", task.Id)
	test.Assertf(t, taskContext.Status == TaskCompleted, "unexpected status %v", taskContext.Status)
	test.Assertf(t, !taskContext.Success, "expecting unsuccessful result")
}

func TestTaskTimeoutKillsLongRunningTask(t *testing.T) {
	task := createSomeGatlingTask()
	tm := NewTaskManager(fakeGatlingOps(5*time.Second), make(chan interface{}, 1024), []uploader.GatlingArtifactUploader{})
	tm.SetTaskTimeout(1 * time.Second)
	resultC, err := tm.SubmitTask(task)
	test.Assertf(t, err == nil, "failed to submit task : %v", err)
	result := <-resultC
	test.Assertf(t, !result.Ok, "expecting task to be killed by timeout")
	taskContext, found := tm.GetTaskRuntimeContext(task.Id)
	test.Assertf(t, found, "task context not found for task %v", task.Id)
	test.Assertf(t, taskContext.Status == TaskAborted, "unexpected status %v", taskContext.Status)
}

func TestNewTaskManagerDefaultsToDefaultTaskTimeout(t *testing.T) {
	tm := NewTaskManager(fakeGatlingOps(0), make(chan interface{}, 1024), []uploader.GatlingArtifactUploader{})
	test.Assertf(t, tm.TaskTimeout == defaultTaskTimeout, "unexpected default timeout %v", tm.TaskTimeout)
}

func TestSetTaskTimeoutOverridesDefault(t *testing.T) {
	tm := NewTaskManager(fakeGatlingOps(0), make(chan interface{}, 1024), []uploader.GatlingArtifactUploader{})
	tm.SetTaskTimeout(5 * time.Minute)
	test.Assertf(t, tm.TaskTimeout == 5*time.Minute, "unexpected timeout %v", tm.TaskTimeout)
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
	snapshot := taskContext.Snapshot()
	test.Assertf(t, snapshot.Status == TaskAborted, "unexpected status %v", snapshot.Status)
}

func createSomeGatlingTask() *gatling.Task {
	tmpDir, _ := os.MkdirTemp("", "")
	userFilesDir, _ := workspace.NewUserFilesDir(filepath.Join(tmpDir, "repos"))
	_ = fileioutil.CopyFile("testdata/SingleFileExampleSimulation.scala",
		fmt.Sprintf("%s/SingleFileExampleSimulation.scala",
			userFilesDir.Simulations))
	return gatling.NewTask(CreateTaskId(), "gatling.test.example.simulation.SingleFileExampleSimulation",
		"-DbaseUrl=http://localhost:8080 -DdurationMin=0.10 -DrequestPerSecond=1", userFilesDir)
}

func fakeGatlingOps(delay time.Duration) gatling.Ops {
	return gatling.RunSimulationFunc(func(_ cmdexec.CommandExecutionOps, task *gatling.Task) (*exec.Cmd, error) {
		cmd := exec.Command("sleep", fmt.Sprintf("%v", delay))
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		return cmd, nil
	})
}

func fakeGatlingOpsWithError(err error) gatling.Ops {
	return gatling.RunSimulationFunc(func(_ cmdexec.CommandExecutionOps, task *gatling.Task) (*exec.Cmd, error) {
		return nil, err
	})
}

// fakeFailingGatlingOps runs a command that starts successfully but exits
// with a non-zero status - not via a signal - to exercise the "real"
// failure branch in worker(), distinct from timeout/abort kills.
func fakeFailingGatlingOps() gatling.Ops {
	return gatling.RunSimulationFunc(func(_ cmdexec.CommandExecutionOps, task *gatling.Task) (*exec.Cmd, error) {
		cmd := exec.Command("false")
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		return cmd, nil
	})
}
