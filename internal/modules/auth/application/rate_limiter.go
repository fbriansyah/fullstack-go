package application

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter defines the interface for rate limiting functionality
type RateLimiter interface {
	// Allow checks if the request is allowed for the given key
	Allow(ctx context.Context, key string) (bool, error)

	// Reset resets the rate limit for the given key
	Reset(ctx context.Context, key string) error

	// GetAttempts returns the current number of attempts for the given key
	GetAttempts(ctx context.Context, key string) (int, error)
}

// InMemoryRateLimiter implements RateLimiter using in-memory storage
type InMemoryRateLimiter struct {
	attempts map[string]*attemptRecord
	mutex    sync.RWMutex
	config   RateLimiterConfig
}

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	MaxAttempts int           // Maximum number of attempts allowed
	Window      time.Duration // Time window for rate limiting
	LockoutTime time.Duration // How long to lock out after max attempts exceeded
}

// attemptRecord tracks attempts for a specific key
type attemptRecord struct {
	count        int
	firstAttempt time.Time
	lastAttempt  time.Time
	lockedUntil  *time.Time
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter
func NewInMemoryRateLimiter(config RateLimiterConfig) *InMemoryRateLimiter {
	limiter := &InMemoryRateLimiter{
		attempts: make(map[string]*attemptRecord),
		config:   config,
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// Allow checks if the request is allowed for the given key
func (r *InMemoryRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	now := time.Now()
	record, exists := r.attempts[key]

	if !exists {
		// First attempt for this key
		r.attempts[key] = &attemptRecord{
			count:        1,
			firstAttempt: now,
			lastAttempt:  now,
		}
		return true, nil
	}

	// Check if currently locked out
	if record.lockedUntil != nil {
		if now.Before(*record.lockedUntil) {
			return false, NewRateLimitExceededError(
				fmt.Sprintf("Too many attempts. Try again after %v",
					record.lockedUntil.Sub(now).Round(time.Second)))
		} else {
			// Lockout has expired, reset the record
			record.count = 1
			record.firstAttempt = now
			record.lastAttempt = now
			record.lockedUntil = nil
			return true, nil
		}
	}

	// Check if window has expired
	if now.Sub(record.firstAttempt) > r.config.Window {
		// Reset the window
		record.count = 1
		record.firstAttempt = now
		record.lastAttempt = now
		record.lockedUntil = nil
		return true, nil
	}

	// Increment attempt count
	record.count++
	record.lastAttempt = now

	// Check if max attempts exceeded
	if record.count > r.config.MaxAttempts {
		// Lock out the key
		lockoutUntil := now.Add(r.config.LockoutTime)
		record.lockedUntil = &lockoutUntil

		return false, NewRateLimitExceededError(
			fmt.Sprintf("Too many attempts. Try again after %v",
				r.config.LockoutTime.Round(time.Second)))
	}

	return true, nil
}

// Reset resets the rate limit for the given key
func (r *InMemoryRateLimiter) Reset(ctx context.Context, key string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.attempts, key)
	return nil
}

// GetAttempts returns the current number of attempts for the given key
func (r *InMemoryRateLimiter) GetAttempts(ctx context.Context, key string) (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	record, exists := r.attempts[key]
	if !exists {
		return 0, nil
	}

	return record.count, nil
}

// cleanup removes expired entries from the attempts map
func (r *InMemoryRateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute * 5) // Cleanup every 5 minutes
	defer ticker.Stop()

	for range ticker.C {
		r.mutex.Lock()
		now := time.Now()

		for key, record := range r.attempts {
			// Remove entries that are older than window + lockout time
			maxAge := r.config.Window + r.config.LockoutTime
			if now.Sub(record.lastAttempt) > maxAge {
				delete(r.attempts, key)
			}
		}

		r.mutex.Unlock()
	}
}

// DefaultRateLimiterConfig returns a default rate limiter configuration
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		MaxAttempts: 5,                // 5 attempts
		Window:      time.Minute * 15, // within 15 minutes
		LockoutTime: time.Minute * 30, // locked out for 30 minutes
	}
}

// LoginRateLimiterConfig returns a rate limiter configuration for login attempts
func LoginRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		MaxAttempts: 5,                // 5 login attempts
		Window:      time.Minute * 15, // within 15 minutes
		LockoutTime: time.Minute * 30, // locked out for 30 minutes
	}
}

// RegistrationRateLimiterConfig returns a rate limiter configuration for registration attempts
func RegistrationRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		MaxAttempts: 3,             // 3 registration attempts
		Window:      time.Hour,     // within 1 hour
		LockoutTime: time.Hour * 2, // locked out for 2 hours
	}
}
