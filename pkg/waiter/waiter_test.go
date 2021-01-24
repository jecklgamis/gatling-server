package waiter

import (
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWaitUntilCountExceed(t *testing.T) {
	var numInvocations = 0
	WaitUntil(100*time.Millisecond, 3, func(counter int) bool {
		numInvocations++
		return false
	})
	test.Assert(t, numInvocations == 3, "unexpected number of invocations")
}

func TestWaitUntilCallbackReturnsTrue(t *testing.T) {
	numInvocations := 0
	WaitUntil(100*time.Millisecond, 3, func(counter int) bool {
		numInvocations++
		return true
	})
	test.Assert(t, numInvocations == 1, "unexpected number of invocations")
}

func TestWaitUntilHTTPGetOk(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(okHandler)
	err := WaitUntilHTTPGetOk(server.URL, 1*time.Second, 3)
	test.Assertf(t, err == nil, "expecting wait to succeed")
}
