package event

import (
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"testing"
	"time"
)

func TestSubmitEvent(t *testing.T) {
	eventBus := NewEventBus()
	taskId := "some-task-id"
	eventBus.EventC <- NewHeartbeatEvent()
	eventBus.EventC <- NewTaskSubmittedEvent()
	eventBus.EventC <- NewTaskStartedEvent(taskId)
	eventBus.EventC <- NewTaskAbortedEvent(taskId)
	eventBus.EventC <- NewTaskCompletedEvent(taskId, true)
	time.Sleep(1 * time.Second)
	test.Assertf(t, eventBus.NumEvents == 5, "invalid number of events received %v", eventBus.NumEvents)
	eventBus.Stop()
}

func TestRegisterListener(t *testing.T) {
	eventBus := NewEventBus()
	counter := 0
	done := make(chan bool)
	eventBus.EventC <- NewHeartbeatEvent()
	eventBus.RegisterListener(ListenerFunc(func(event interface{}) {
		counter++
		done <- true
	}))
	<-done
	test.Assertf(t, counter == 1, "invalid number of events received %v", counter)
	eventBus.Stop()
}
