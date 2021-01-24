package waiter

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func WaitUntil(retryDelay time.Duration, retries int, callback func(counter int) bool) error {
	var exitNow = false
	counter := 1
	for counter <= retries && !exitNow {
		if callback(counter) {
			exitNow = true
		}
		counter++
		time.Sleep(retryDelay)
	}
	if counter > retries {
		return fmt.Errorf("gave up waiting after %d attempts", retries)
	}
	return nil
}

func WaitUntilHTTPGetOk(url string, delay time.Duration, retries int) error {
	log.Printf("Waiting until GET %s is OK (delay=%s, retries=%d)\n",
		url, delay.String(), retries)
	err := WaitUntil(delay, retries, func(counter int) bool {
		resp, err := http.Get(url)
		if err != nil {
			return false
		}
		if resp.StatusCode == http.StatusOK {
			log.Println("OK", url)
		}
		return resp.StatusCode == http.StatusOK
	})
	if err != nil {
		log.Println("!OK", url)
	}
	return err
}
