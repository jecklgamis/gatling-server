package event

import (
	"github.com/jecklgamis/gatling-server/pkg/jsonutil"
)

type TaskCompletedEvent struct {
	Name     string            `json:"name"`
	TaskId   string            `json:"taskId"`
	Ok       bool              `json:"ok"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func (e *TaskCompletedEvent) String() string {
	return jsonutil.ToJson(e)
}

type TaskStartedEvent struct {
	Name     string            `json:"name"`
	TaskId   string            `json:"taskId"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func (e *TaskStartedEvent) String() string {
	return jsonutil.ToJson(e)
}

type TaskAbortedEvent struct {
	Name     string            `json:"name"`
	TaskId   string            `json:"taskId"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func (e *TaskAbortedEvent) String() string {
	return jsonutil.ToJson(e)
}

type TaskSubmittedEvent struct {
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func (e *TaskSubmittedEvent) String() string {
	return jsonutil.ToJson(e)
}

type HeartbeatEvent struct {
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func (e *HeartbeatEvent) String() string {
	return jsonutil.ToJson(e)
}

func NewHeartbeatEvent() *HeartbeatEvent {
	return &HeartbeatEvent{Name: "HeartbeatEvent", Metadata: emptyMetadata()}
}

func NewTaskCompletedEvent(taskId string, ok bool) *TaskCompletedEvent {
	return &TaskCompletedEvent{Name: "TaskCompletedEvent", TaskId: taskId, Ok: ok, Metadata: emptyMetadata()}
}

func NewTaskSubmittedEvent() *TaskSubmittedEvent {
	return &TaskSubmittedEvent{Name: "TaskSubmittedEvent", Metadata: emptyMetadata()}
}

func NewTaskAbortedEvent(taskId string) *TaskAbortedEvent {
	return &TaskAbortedEvent{Name: "TaskAbortedEvent", TaskId: taskId, Metadata: emptyMetadata()}
}

func NewTaskStartedEvent(taskId string) *TaskStartedEvent {
	return &TaskStartedEvent{Name: "TaskStartedEvent", TaskId: taskId, Metadata: emptyMetadata()}
}

func emptyMetadata() map[string]string {
	return map[string]string{}
}
