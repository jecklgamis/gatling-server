package taskmanager

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/jecklgamis/gatling-server/pkg/cmdexec"
	"github.com/jecklgamis/gatling-server/pkg/event"
	"github.com/jecklgamis/gatling-server/pkg/gatling"
	"github.com/jecklgamis/gatling-server/pkg/tarutil"
	"github.com/jecklgamis/gatling-server/pkg/uploader"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type TaskStatus string

const (
	TaskStarted   TaskStatus = "Started"
	TaskAborted   TaskStatus = "Aborted"
	TaskCompleted TaskStatus = "Completed"
)

type TaskRuntimeContext struct {
	Process   *os.Process
	Task      *gatling.Task
	Started   time.Time
	Completed time.Time
	Duration  time.Duration
	Status    TaskStatus
	Success   bool
}

func (c *TaskRuntimeContext) markStarted() {
	c.Started = time.Now()
	c.Status = TaskStarted
}

func (c *TaskRuntimeContext) markCompleted(success bool) {
	c.Completed = time.Now()
	c.Status = TaskCompleted
	c.Duration = c.Completed.Sub(c.Started)
	c.Success = success
}

func (c *TaskRuntimeContext) markAborted() {
	c.Completed = time.Now()
	c.Status = TaskAborted
	c.Duration = time.Now().Sub(c.Started)
	c.Success = false
}

type Ops interface {
	SubmitTask(task *gatling.Task) (chan *gatling.Result, error)
	AbortTask(taskId string) error
	GetTaskRuntimeContext(taskId string) (*TaskRuntimeContext, bool)
}

type TaskManager struct {
	taskContextMutex  sync.Mutex
	TaskContexts      map[string]*TaskRuntimeContext
	Gatling           gatling.Ops
	EventChannel      chan interface{}
	artifactUploaders []uploader.GatlingArtifactUploader
}

func CreateTaskId() string {
	return uuid.New().String()[0:8]
}

func NewTaskManager(gatlingOps gatling.Ops, eventChannel chan interface{},
	artifactUploaders []uploader.GatlingArtifactUploader) *TaskManager {
	return &TaskManager{Gatling: gatlingOps, EventChannel: eventChannel, artifactUploaders: artifactUploaders,
		TaskContexts: map[string]*TaskRuntimeContext{}}
}

func (t *TaskManager) GetTaskRuntimeContext(taskId string) (*TaskRuntimeContext, bool) {
	context, ok := t.TaskContexts[taskId]
	if !ok {
		return nil, false
	}
	return context, true
}

func (t *TaskManager) AbortTask(taskId string) error {
	context, ok := t.TaskContexts[taskId]
	if !ok {
		return fmt.Errorf("task %v not found", taskId)
	}
	if context.Status == TaskCompleted {
		return fmt.Errorf("task %v already completed", taskId)
	}
	if context.Status == TaskAborted {
		return fmt.Errorf("task %v already aborted", taskId)
	}
	if context.Process != nil {
		defer context.markAborted()
		err := context.Process.Kill()
		if err != nil {
			return err
		}
		log.Println("Triggered abort on task", taskId)
		return nil
	} else {
		return fmt.Errorf("task process is not set")
	}
}

func (t *TaskManager) worker(context *TaskRuntimeContext, task *gatling.Task, result chan<- *gatling.Result) {
	log.Printf("Gatling task %v started", task.Id)
	t.EventChannel <- event.NewTaskStartedEvent(task.Id)
	defer func() {
		tarutil.CompressDir(task.UserFilesDir.Results, task.UserFilesDir.BaseDir, "results.tar.gz")
		for _, uploader := range t.artifactUploaders {
			uploader.Upload(task.Id, task.UserFilesDir)
		}
	}()
	context.markStarted()
	cmd, err := t.Gatling.RunSimulation(cmdexec.NewCommandExecutor(), task)
	if err != nil {
		log.Println("Failed executing command :", err)
		context.markCompleted(false)
		result <- &gatling.Result{Ok: false}
		return
	}
	context.Process = cmd.Process
	log.Println("Waiting for task", task.Id, "to complete")
	err = cmd.Wait()
	if err != nil {
		log.Println("Failed executing command :", err)
		if strings.Contains(err.Error(), "signal: killed") {
			context.markAborted()
			result <- &gatling.Result{Ok: false}
			t.EventChannel <- event.NewTaskAbortedEvent(task.Id)
		} else {
			context.markCompleted(false)
			result <- &gatling.Result{Ok: false}
			t.EventChannel <- event.NewTaskCompletedEvent(task.Id, false)
		}
		return
	}
	context.markCompleted(true)
	result <- &gatling.Result{Ok: true}
	t.EventChannel <- event.NewTaskCompletedEvent(task.Id, true)
}

func (t *TaskManager) SubmitTask(task *gatling.Task) (chan *gatling.Result, error) {
	if task == nil {
		return nil, fmt.Errorf("task is nil")
	}
	context := TaskRuntimeContext{Task: task}
	t.storeContext(&context)
	result := make(chan *gatling.Result, 1)
	go t.worker(&context, task, result)
	return result, nil
}

func (t *TaskManager) storeContext(c *TaskRuntimeContext) {
	t.taskContextMutex.Lock()
	t.TaskContexts[c.Task.Id] = c
	t.taskContextMutex.Unlock()
}
