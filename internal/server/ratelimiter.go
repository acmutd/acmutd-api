package server

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu     sync.Mutex
	limits map[string]*rateLimit
}

type rateLimit struct {
	count     int
	windowEnd time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limits: make(map[string]*rateLimit),
	}
}

func (rl *RateLimiter) Allow(key string, limit int, windowSeconds int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowDuration := time.Duration(windowSeconds) * time.Second

	// Initialize or reset rate limit window
	if rl.limits[key] == nil || now.After(rl.limits[key].windowEnd) {
		rl.limits[key] = &rateLimit{
			count:     1,
			windowEnd: now.Add(windowDuration),
		}
		return true
	}

	// Check if within limit
	if rl.limits[key].count < limit {
		rl.limits[key].count++
		return true
	}

	return false
}

func (rl *RateLimiter) StartCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			rl.mu.Lock()
			now := time.Now()
			for key, limit := range rl.limits {
				if now.After(limit.windowEnd.Add(5 * time.Minute)) {
					delete(rl.limits, key)
				}
			}
			rl.mu.Unlock()
		}
	}()
}
