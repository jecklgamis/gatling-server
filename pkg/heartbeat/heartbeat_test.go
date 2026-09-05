package heartbeat

import (
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"testing"
	"time"
)

func TestStartHeartBeat(t *testing.T) {
	var counter = 0
	received := make(chan bool)
	heartbeat, err := New(1*time.Second, func() {
		counter++
		received <- true
	})
	test.Assertf(t, err == nil, "unexpected error :%v", err)
	<-received
	heartbeat.Stop()
	heartbeat.Stop()
	test.Assertf(t, counter == 1, "unexpected count %v", counter)
}

func TestNewHeartBeatRejectsNonPositiveFrequency(t *testing.T) {
	_, err := New(0, func() {})
	test.Assertf(t, err != nil, "expecting error for non-positive frequency")
}
