package ratelimit

import (
	"sync"
	"time"
)

// Limiter tracks request counts in fixed windows per key.
type Limiter struct {
	mu     sync.Mutex
	limits map[string]*window
}

type window struct {
	count     int
	windowEnd time.Time
}

// NewLimiter creates a limiter with in-memory tracking.
func NewLimiter() *Limiter {
	return &Limiter{
		limits: make(map[string]*window),
	}
}

// Allow returns true if the request is within the configured limit.
func (l *Limiter) Allow(key string, limit int, windowSeconds int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	windowDuration := time.Duration(windowSeconds) * time.Second

	win := l.limits[key]
	if win == nil || now.After(win.windowEnd) {
		l.limits[key] = &window{
			count:     1,
			windowEnd: now.Add(windowDuration),
		}
		return true
	}

	if win.count < limit {
		win.count++
		return true
	}

	return false
}

// StartCleanup periodically evicts stale windows to limit memory usage.
func (l *Limiter) StartCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			l.mu.Lock()
			now := time.Now()
			for key, win := range l.limits {
				if now.After(win.windowEnd.Add(5 * time.Minute)) {
					delete(l.limits, key)
				}
			}
			l.mu.Unlock()
		}
	}()
}
