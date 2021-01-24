package event

import (
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"net/http"
	"testing"
)

func TestNoUrlInConfig(t *testing.T) {
	kv := map[string]string{}
	test.Assertf(t, NewHTTPNotifier(kv) == nil, "expecting nil HTTP notifier")
}

func TestValidConfig(t *testing.T) {
	kv := map[string]string{"url": "http://localhost:8080"}
	test.Assertf(t, NewHTTPNotifier(kv) != nil, "expecting valid HTTP notifier")
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func internalServerErrorHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
}

func TestPostEvent(t *testing.T) {
	server := NewMinimalHttpServer()
	defer server.close()
	server.handle("/", okHandler())
	notifier := &HttpEventNotifier{map[string]string{"url": server.URL}}
	err := notifier.notify(NewHeartbeatEvent())
	test.Assertf(t, err == nil, "failed sending event : %v", err)
}

func TestPostEventAnd5xxResponse(t *testing.T) {
	server := NewMinimalHttpServer()
	defer server.close()
	server.handle("/", internalServerErrorHandler())
	notifier := &HttpEventNotifier{map[string]string{"url": server.URL}}
	err := notifier.notify(NewHeartbeatEvent())
	test.Assert(t, err != nil, "expecting error response")
}

func TestPostEventOnUnknownHost(t *testing.T) {
	notifier := &HttpEventNotifier{map[string]string{"url": "http://some-host"}}
	err := notifier.notify(NewHeartbeatEvent())
	test.Assert(t, err != nil, "expecting error response")
}
