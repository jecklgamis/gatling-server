package handler

import (
	"sync"
	"time"
)

const (
	authMaxFailures = 5
	authWindow      = time.Minute
)

type authAttempt struct {
	count     int
	firstSeen time.Time
}

type authLimiter struct {
	mu       sync.Mutex
	attempts map[string]*authAttempt
}

func newAuthLimiter() *authLimiter {
	return &authLimiter{attempts: make(map[string]*authAttempt)}
}

func (l *authLimiter) blocked(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	attempt, ok := l.attempts[key]
	if !ok {
		return false
	}
	if time.Since(attempt.firstSeen) > authWindow {
		delete(l.attempts, key)
		return false
	}
	return attempt.count >= authMaxFailures
}

func (l *authLimiter) recordFailure(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	attempt, ok := l.attempts[key]
	now := time.Now()
	if !ok || now.Sub(attempt.firstSeen) > authWindow {
		l.attempts[key] = &authAttempt{count: 1, firstSeen: now}
		return
	}
	attempt.count++
}
