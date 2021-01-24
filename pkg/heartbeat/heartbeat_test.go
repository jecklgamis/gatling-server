package heartbeat

import (
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"testing"
	"time"
)

func TestStartHeartBeat(t *testing.T) {
	var counter = 0
	received := make(chan bool)
	heartbeat := New(1*time.Second, func() {
		counter++
		received <- true
	})
	<-received
	heartbeat.Stop()
	test.Assertf(t, counter == 1, "unexpected count %v", counter)
}
