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
	"sync/atomic"
	"time"
)

const defaultTaskTimeout = 30 * time.Minute

type TaskStatus string

const (
	TaskStarted   TaskStatus = "Started"
	TaskAborted   TaskStatus = "Aborted"
	TaskCompleted TaskStatus = "Completed"
)

type TaskRuntimeContext struct {
	mu        sync.Mutex
	Process   *os.Process
	Task      *gatling.Task
	Started   time.Time
	Completed time.Time
	Duration  time.Duration
	Status    TaskStatus
	Success   bool
}

func (c *TaskRuntimeContext) markStarted() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Started = time.Now()
	c.Status = TaskStarted
}

func (c *TaskRuntimeContext) markCompleted(success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Completed = time.Now()
	c.Status = TaskCompleted
	c.Duration = c.Completed.Sub(c.Started)
	c.Success = success
}

func (c *TaskRuntimeContext) markAborted() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Completed = time.Now()
	c.Status = TaskAborted
	c.Duration = time.Now().Sub(c.Started)
	c.Success = false
}

func (c *TaskRuntimeContext) setProcess(p *os.Process) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Process = p
}

func (c *TaskRuntimeContext) getProcess() *os.Process {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Process
}

func (c *TaskRuntimeContext) getStatus() TaskStatus {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Status
}

// Snapshot returns a point-in-time copy of the context's fields, safe to read
// or JSON-marshal without racing the worker goroutine that mutates them.
func (c *TaskRuntimeContext) Snapshot() TaskRuntimeContext {
	c.mu.Lock()
	defer c.mu.Unlock()
	return TaskRuntimeContext{
		Process:   c.Process,
		Task:      c.Task,
		Started:   c.Started,
		Completed: c.Completed,
		Duration:  c.Duration,
		Status:    c.Status,
		Success:   c.Success,
	}
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
	TaskTimeout       time.Duration
}

func CreateTaskId() string {
	return uuid.New().String()
}

func NewTaskManager(gatlingOps gatling.Ops, eventChannel chan interface{},
	artifactUploaders []uploader.GatlingArtifactUploader) *TaskManager {
	return &TaskManager{Gatling: gatlingOps, EventChannel: eventChannel, artifactUploaders: artifactUploaders,
		TaskContexts: map[string]*TaskRuntimeContext{}, TaskTimeout: defaultTaskTimeout}
}

// SetTaskTimeout overrides the default duration a simulation is allowed to run
// before it is forcibly killed. Pass 0 to disable the timeout.
func (t *TaskManager) SetTaskTimeout(timeout time.Duration) {
	t.TaskTimeout = timeout
}

func (t *TaskManager) GetTaskRuntimeContext(taskId string) (*TaskRuntimeContext, bool) {
	t.taskContextMutex.Lock()
	defer t.taskContextMutex.Unlock()
	context, ok := t.TaskContexts[taskId]
	if !ok {
		return nil, false
	}
	return context, true
}

func (t *TaskManager) AbortTask(taskId string) error {
	t.taskContextMutex.Lock()
	context, ok := t.TaskContexts[taskId]
	t.taskContextMutex.Unlock()
	if !ok {
		return fmt.Errorf("task %v not found", taskId)
	}
	status := context.getStatus()
	if status == TaskCompleted {
		return fmt.Errorf("task %v already completed", taskId)
	}
	if status == TaskAborted {
		return fmt.Errorf("task %v already aborted", taskId)
	}
	if process := context.getProcess(); process != nil {
		defer context.markAborted()
		err := process.Kill()
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
	context.setProcess(cmd.Process)
	log.Println("Waiting for task", task.Id, "to complete")
	var timedOut int32
	var timer *time.Timer
	if t.TaskTimeout > 0 {
		timer = time.AfterFunc(t.TaskTimeout, func() {
			atomic.StoreInt32(&timedOut, 1)
			log.Println("Task", task.Id, "exceeded timeout of", t.TaskTimeout, "- killing process")
			if err := cmd.Process.Kill(); err != nil {
				log.Println("Unable to kill timed-out task", task.Id, ":", err)
			}
		})
	}
	err = cmd.Wait()
	if timer != nil {
		timer.Stop()
	}
	if err != nil {
		log.Println("Failed executing command :", err)
		if atomic.LoadInt32(&timedOut) == 1 {
			log.Println("Task", task.Id, "was killed after exceeding timeout of", t.TaskTimeout)
			context.markAborted()
			result <- &gatling.Result{Ok: false}
			t.EventChannel <- event.NewTaskAbortedEvent(task.Id)
		} else if strings.Contains(err.Error(), "signal: killed") {
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
