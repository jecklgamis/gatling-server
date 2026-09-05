package heartbeat

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type HeartBeat struct {
	ticker   *time.Ticker
	done     chan struct{}
	stopOnce sync.Once
}

func New(frequency time.Duration, callback func()) (*HeartBeat, error) {
	if frequency <= 0 {
		return nil, fmt.Errorf("heartbeat frequency must be positive, got %v", frequency)
	}
	heartbeat := &HeartBeat{
		ticker: time.NewTicker(frequency),
		done:   make(chan struct{}),
	}
	go func() {
		defer heartbeat.ticker.Stop()
		for {
			select {
			case <-heartbeat.done:
				log.Println("Heartbeat stopped")
				return
			case <-heartbeat.ticker.C:
				go safeCall(callback)
			}
		}
	}()
	return heartbeat, nil
}

// safeCall runs callback, recovering from any panic so a misbehaving callback
// cannot bring down the whole process from a detached goroutine.
func safeCall(callback func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Heartbeat callback panicked :", r)
		}
	}()
	callback()
}

// Stop terminates the heartbeat. Safe to call more than once or concurrently.
func (r *HeartBeat) Stop() {
	r.stopOnce.Do(func() {
		close(r.done)
	})
}
