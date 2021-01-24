package event

import "log"

type Listener interface {
	Event(event interface{})
}

type ListenerFunc func(event interface{})

func (l ListenerFunc) Event(event interface{}) {
	l(event)
}

type Bus struct {
	EventC    chan interface{}
	StopC     chan bool
	NumEvents uint64
	listeners []Listener
}

func NewEventBus() *Bus {
	notifier := &Bus{EventC: make(chan interface{}, 1024), StopC: make(chan bool, 1),
		listeners: make([]Listener, 0)}
	go notifier.selectChannelEvents()
	return notifier
}

func (b *Bus) RegisterListener(listener Listener) {
	b.listeners = append(b.listeners, listener)
}

func (b *Bus) selectChannelEvents() {
	for {
		select {
		case <-b.StopC:
			log.Println("Event bus stopped")
			return
		case e := <-b.EventC:
			log.Println("Event published", e)
			b.NumEvents++
			b.handleEvent(e)

		}
	}
}

func (b *Bus) handleEvent(event interface{}) {
	switch e := event.(type) {
	case *HeartbeatEvent, *TaskSubmittedEvent, *TaskStartedEvent, *TaskAbortedEvent, *TaskCompletedEvent:
		for _, l := range b.listeners {
			go l.Event(e)
		}
	default:
		log.Println("Unknown event", e)
	}
}

func (b *Bus) Stop() {
	b.StopC <- true
	close(b.EventC)
	close(b.StopC)
}
