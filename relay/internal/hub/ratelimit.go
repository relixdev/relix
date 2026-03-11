package hub

import (
	"sync"
	"time"
)

// DefaultRateLimitMaxAttempts is the default max pairing attempts per window.
const DefaultRateLimitMaxAttempts = 5

// DefaultRateLimitWindow is the default rate-limit window duration.
const DefaultRateLimitWindow = 3 * time.Second

// rateLimitOptions configures a RateLimiter.
type rateLimitOptions struct {
	maxAttempts int
	window      time.Duration
}

// ipBucket tracks attempt count and window expiry for a single IP.
type ipBucket struct {
	count     int
	windowEnd time.Time
}

// RateLimiter tracks per-IP attempt counts within a fixed window.
// It is safe for concurrent use.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*ipBucket
	opts    rateLimitOptions
}

// newRateLimiter creates a RateLimiter with the given options (package-internal).
func newRateLimiter(opts rateLimitOptions) *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*ipBucket),
		opts:    opts,
	}
}

// NewRateLimiter creates a RateLimiter with explicit maxAttempts and window.
func NewRateLimiter(maxAttempts int, window time.Duration) *RateLimiter {
	return newRateLimiter(rateLimitOptions{
		maxAttempts: maxAttempts,
		window:      window,
	})
}

// newDefaultRateLimiter creates a RateLimiter with production defaults.
func newDefaultRateLimiter() *RateLimiter {
	return NewRateLimiter(DefaultRateLimitMaxAttempts, DefaultRateLimitWindow)
}

// Allow returns true if the given IP is within its rate limit, false if blocked.
// Each call that returns true counts as one attempt.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b := rl.buckets[ip]
	if b == nil || now.After(b.windowEnd) {
		// New or expired bucket — start fresh window.
		rl.buckets[ip] = &ipBucket{
			count:     1,
			windowEnd: now.Add(rl.opts.window),
		}
		return true
	}

	if b.count >= rl.opts.maxAttempts {
		return false
	}
	b.count++
	return true
}
