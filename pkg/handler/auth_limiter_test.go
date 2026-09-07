package handler

import (
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"testing"
	"time"
)

func TestNotBlockedInitially(t *testing.T) {
	l := newAuthLimiter()
	test.Assertf(t, !l.blocked("some-key"), "expecting not blocked")
}

func TestNotBlockedBelowMaxFailures(t *testing.T) {
	l := newAuthLimiter()
	for i := 0; i < authMaxFailures-1; i++ {
		l.recordFailure("some-key")
	}
	test.Assertf(t, !l.blocked("some-key"), "expecting not blocked below max failures")
}

func TestBlockedAtMaxFailures(t *testing.T) {
	l := newAuthLimiter()
	for i := 0; i < authMaxFailures; i++ {
		l.recordFailure("some-key")
	}
	test.Assertf(t, l.blocked("some-key"), "expecting blocked at max failures")
}

func TestDifferentKeysTrackedIndependently(t *testing.T) {
	l := newAuthLimiter()
	for i := 0; i < authMaxFailures; i++ {
		l.recordFailure("key-a")
	}
	test.Assertf(t, l.blocked("key-a"), "expecting key-a blocked")
	test.Assertf(t, !l.blocked("key-b"), "expecting key-b unaffected")
}

func TestBlockClearsAfterWindowExpires(t *testing.T) {
	l := newAuthLimiter()
	l.attempts["some-key"] = &authAttempt{count: authMaxFailures, firstSeen: time.Now().Add(-2 * authWindow)}
	test.Assertf(t, !l.blocked("some-key"), "expecting expired window to no longer be blocked")
	_, found := l.attempts["some-key"]
	test.Assertf(t, !found, "expecting expired entry to be removed")
}

func TestRecordFailureResetsAfterWindowExpires(t *testing.T) {
	l := newAuthLimiter()
	l.attempts["some-key"] = &authAttempt{count: authMaxFailures, firstSeen: time.Now().Add(-2 * authWindow)}
	l.recordFailure("some-key")
	test.Assertf(t, l.attempts["some-key"].count == 1, "expecting count reset to 1, got %d", l.attempts["some-key"].count)
	test.Assertf(t, !l.blocked("some-key"), "expecting not blocked after reset")
}
