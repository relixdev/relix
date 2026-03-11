package hub

import (
	"testing"
	"time"
)

func TestRateLimiterAllowsFiveAttempts(t *testing.T) {
	rl := newRateLimiter(rateLimitOptions{
		maxAttempts: 5,
		window:      3 * time.Second,
	})

	for i := 0; i < 5; i++ {
		if !rl.Allow("1.2.3.4") {
			t.Errorf("attempt %d: want allowed, got blocked", i+1)
		}
	}
}

func TestRateLimiterBlocksSixthAttempt(t *testing.T) {
	rl := newRateLimiter(rateLimitOptions{
		maxAttempts: 5,
		window:      3 * time.Second,
	})

	for i := 0; i < 5; i++ {
		rl.Allow("1.2.3.4")
	}

	if rl.Allow("1.2.3.4") {
		t.Error("6th attempt: want blocked, got allowed")
	}
}

func TestRateLimiterResetsAfterWindow(t *testing.T) {
	window := 60 * time.Millisecond
	rl := newRateLimiter(rateLimitOptions{
		maxAttempts: 2,
		window:      window,
	})

	rl.Allow("1.2.3.4")
	rl.Allow("1.2.3.4")

	// Now blocked.
	if rl.Allow("1.2.3.4") {
		t.Fatal("should be blocked before window expires")
	}

	// Wait for window to expire.
	time.Sleep(window + 20*time.Millisecond)

	if !rl.Allow("1.2.3.4") {
		t.Error("want allowed after window reset, got blocked")
	}
}

func TestRateLimiterIsolatesIPs(t *testing.T) {
	rl := newRateLimiter(rateLimitOptions{
		maxAttempts: 2,
		window:      3 * time.Second,
	})

	rl.Allow("1.1.1.1")
	rl.Allow("1.1.1.1")

	// 1.1.1.1 is now blocked.
	if rl.Allow("1.1.1.1") {
		t.Error("1.1.1.1: should be blocked")
	}

	// 2.2.2.2 has its own counter — should still be allowed.
	if !rl.Allow("2.2.2.2") {
		t.Error("2.2.2.2: should not be affected by 1.1.1.1 limit")
	}
}

func TestRateLimiterDefaultOptions(t *testing.T) {
	// Verify the default constructor uses the correct production limits:
	// max 5 attempts per 3-second window.
	rl := newDefaultRateLimiter()

	for i := 0; i < 5; i++ {
		if !rl.Allow("10.0.0.1") {
			t.Errorf("attempt %d: want allowed", i+1)
		}
	}
	if rl.Allow("10.0.0.1") {
		t.Error("6th attempt with default limiter: want blocked")
	}
}
