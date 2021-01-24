package testing

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

func nextPort() int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Intn(65535-1025) + 1025
}

var mutex = sync.Mutex{}

func UnusedPort() int {
	var portAvailable = false
	dialer := net.Dialer{Timeout: 2 * time.Second}
	for {
		mutex.Lock()
		var port = nextPort()
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		conn, err := dialer.Dial("tcp", addr)
		if err != nil {
			portAvailable = true
		} else {
			_ = conn.Close()
		}
		if portAvailable {
			mutex.Unlock()
			return port
		}
	}
}
