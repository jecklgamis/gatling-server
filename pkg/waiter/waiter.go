package waiter

import (
	"fmt"
	"log/slog"
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
	return WaitUntilHTTPGetOkWithHeaders(url, nil, delay, retries)
}

func WaitUntilHTTPGetOkWithHeaders(url string, headers map[string]string, delay time.Duration, retries int) error {
	slog.Info("Waiting until GET is OK", "url", url, "delay", delay.String(), "retries", retries)
	err := WaitUntil(delay, retries, func(counter int) bool {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return false
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return false
		}
		if resp.StatusCode == http.StatusOK {
			slog.Info("OK", "url", url)
		}
		return resp.StatusCode == http.StatusOK
	})
	if err != nil {
		slog.Info("!OK", "url", url)
	}
	return err
}
