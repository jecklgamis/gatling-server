package heartbeat

import (
	"log"
	"time"
)

type HeartBeat struct {
	frequency time.Duration
	ticker    *time.Ticker
	done      chan bool
}

func New(frequency time.Duration, callback func()) *HeartBeat {
	heartbeat := &HeartBeat{frequency: frequency,
		ticker: time.NewTicker(frequency),
		done:   make(chan bool)}
	go func() {
		for {
			select {
			case <-heartbeat.done:
				log.Println("Heartbeat stopped")
				return
			case _ = <-heartbeat.ticker.C:
				callback()
			}
		}
	}()
	return heartbeat
}

func (r *HeartBeat) Stop() {
	r.done <- true
}
