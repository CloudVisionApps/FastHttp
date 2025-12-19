package ratelimit

import (
	"sync"
	"time"
)

type Limiter struct {
	requests    map[string][]time.Time
	mu          sync.RWMutex
	maxRequests int
	window      time.Duration
}

func New(maxRequests int, windowSeconds int) *Limiter {
	rl := &Limiter{
		requests:    make(map[string][]time.Time),
		maxRequests: maxRequests,
		window:      time.Duration(windowSeconds) * time.Second,
	}

	// Cleanup old entries periodically
	go rl.cleanup()

	return rl
}

func (rl *Limiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, timestamps := range rl.requests {
			validTimestamps := []time.Time{}
			for _, ts := range timestamps {
				if now.Sub(ts) < rl.window {
					validTimestamps = append(validTimestamps, ts)
				}
			}
			if len(validTimestamps) == 0 {
				delete(rl.requests, ip)
			} else {
				rl.requests[ip] = validTimestamps
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *Limiter) Allow(ip string) bool {
	if rl.maxRequests <= 0 {
		return true // Rate limiting disabled
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Clean old timestamps for this IP
	timestamps := rl.requests[ip]
	validTimestamps := []time.Time{}
	for _, ts := range timestamps {
		if now.Sub(ts) < rl.window {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Check if limit exceeded
	if len(validTimestamps) >= rl.maxRequests {
		return false
	}

	// Add current request
	validTimestamps = append(validTimestamps, now)
	rl.requests[ip] = validTimestamps

	return true
}
