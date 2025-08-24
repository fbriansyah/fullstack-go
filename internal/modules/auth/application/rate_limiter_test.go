package application

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryRateLimiter_Allow_FirstAttempt(t *testing.T) {
	config := RateLimiterConfig{
		MaxAttempts: 3,
		Window:      time.Minute,
		LockoutTime: time.Minute * 5,
	}
	limiter := NewInMemoryRateLimiter(config)
	ctx := context.Background()

	// First attempt should be allowed
	allowed, err := limiter.Allow(ctx, "test-key")
	require.NoError(t, err)
	assert.True(t, allowed)

	// Check attempt count
	attempts, err := limiter.GetAttempts(ctx, "test-key")
	require.NoError(t, err)
	assert.Equal(t, 1, attempts)
}

func TestInMemoryRateLimiter_Allow_WithinLimit(t *testing.T) {
	config := RateLimiterConfig{
		MaxAttempts: 3,
		Window:      time.Minute,
		LockoutTime: time.Minute * 5,
	}
	limiter := NewInMemoryRateLimiter(config)
	ctx := context.Background()

	// Make attempts within limit
	for i := 1; i <= 3; i++ {
		allowed, err := limiter.Allow(ctx, "test-key")
		require.NoError(t, err)
		assert.True(t, allowed)

		attempts, err := limiter.GetAttempts(ctx, "test-key")
		require.NoError(t, err)
		assert.Equal(t, i, attempts)
	}
}

func TestInMemoryRateLimiter_Allow_ExceedsLimit(t *testing.T) {
	config := RateLimiterConfig{
		MaxAttempts: 3,
		Window:      time.Minute,
		LockoutTime: time.Minute * 5,
	}
	limiter := NewInMemoryRateLimiter(config)
	ctx := context.Background()

	// Make attempts up to limit
	for i := 1; i <= 3; i++ {
		allowed, err := limiter.Allow(ctx, "test-key")
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	// Fourth attempt should be blocked
	allowed, err := limiter.Allow(ctx, "test-key")
	assert.Error(t, err)
	assert.False(t, allowed)

	// Check that it's a rate limit error
	rateLimitErr, ok := err.(*AuthError)
	require.True(t, ok)
	assert.Equal(t, ErrorCodeRateLimitExceeded, rateLimitErr.Code)
}

func TestInMemoryRateLimiter_Allow_WindowExpired(t *testing.T) {
	config := RateLimiterConfig{
		MaxAttempts: 2,
		Window:      time.Millisecond * 100, // Very short window for testing
		LockoutTime: time.Minute,
	}
	limiter := NewInMemoryRateLimiter(config)
	ctx := context.Background()

	// Make attempts up to limit
	allowed, err := limiter.Allow(ctx, "test-key")
	require.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = limiter.Allow(ctx, "test-key")
	require.NoError(t, err)
	assert.True(t, allowed)

	// Wait for window to expire
	time.Sleep(time.Millisecond * 150)

	// Should be allowed again after window expires
	allowed, err = limiter.Allow(ctx, "test-key")
	require.NoError(t, err)
	assert.True(t, allowed)

	// Attempt count should be reset
	attempts, err := limiter.GetAttempts(ctx, "test-key")
	require.NoError(t, err)
	assert.Equal(t, 1, attempts)
}

func TestInMemoryRateLimiter_Allow_LockoutPeriod(t *testing.T) {
	config := RateLimiterConfig{
		MaxAttempts: 2,
		Window:      time.Minute,
		LockoutTime: time.Millisecond * 100, // Short lockout for testing
	}
	limiter := NewInMemoryRateLimiter(config)
	ctx := context.Background()

	// Exceed limit to trigger lockout
	limiter.Allow(ctx, "test-key")
	limiter.Allow(ctx, "test-key")

	// This should trigger lockout
	allowed, err := limiter.Allow(ctx, "test-key")
	assert.Error(t, err)
	assert.False(t, allowed)

	// Should still be locked out immediately after
	allowed, err = limiter.Allow(ctx, "test-key")
	assert.Error(t, err)
	assert.False(t, allowed)

	// Wait for lockout to expire
	time.Sleep(time.Millisecond * 200)

	// Should be allowed after lockout expires
	allowed, err = limiter.Allow(ctx, "test-key")
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestInMemoryRateLimiter_Reset(t *testing.T) {
	config := RateLimiterConfig{
		MaxAttempts: 2,
		Window:      time.Minute,
		LockoutTime: time.Minute,
	}
	limiter := NewInMemoryRateLimiter(config)
	ctx := context.Background()

	// Make some attempts
	limiter.Allow(ctx, "test-key")
	limiter.Allow(ctx, "test-key")

	// Verify attempts are recorded
	attempts, err := limiter.GetAttempts(ctx, "test-key")
	require.NoError(t, err)
	assert.Equal(t, 2, attempts)

	// Reset the key
	err = limiter.Reset(ctx, "test-key")
	require.NoError(t, err)

	// Attempts should be reset
	attempts, err = limiter.GetAttempts(ctx, "test-key")
	require.NoError(t, err)
	assert.Equal(t, 0, attempts)

	// Should be allowed again
	allowed, err := limiter.Allow(ctx, "test-key")
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestInMemoryRateLimiter_GetAttempts_NonExistentKey(t *testing.T) {
	config := DefaultRateLimiterConfig()
	limiter := NewInMemoryRateLimiter(config)
	ctx := context.Background()

	// Get attempts for non-existent key
	attempts, err := limiter.GetAttempts(ctx, "non-existent")
	require.NoError(t, err)
	assert.Equal(t, 0, attempts)
}

func TestInMemoryRateLimiter_MultipleKeys(t *testing.T) {
	config := RateLimiterConfig{
		MaxAttempts: 2,
		Window:      time.Minute,
		LockoutTime: time.Minute,
	}
	limiter := NewInMemoryRateLimiter(config)
	ctx := context.Background()

	// Make attempts for different keys
	allowed1, err := limiter.Allow(ctx, "key1")
	require.NoError(t, err)
	assert.True(t, allowed1)

	allowed2, err := limiter.Allow(ctx, "key2")
	require.NoError(t, err)
	assert.True(t, allowed2)

	// Each key should have independent counts
	attempts1, err := limiter.GetAttempts(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, 1, attempts1)

	attempts2, err := limiter.GetAttempts(ctx, "key2")
	require.NoError(t, err)
	assert.Equal(t, 1, attempts2)

	// Exceed limit for key1
	limiter.Allow(ctx, "key1")
	allowed1, err = limiter.Allow(ctx, "key1")
	assert.Error(t, err)
	assert.False(t, allowed1)

	// key2 should still be allowed
	allowed2, err = limiter.Allow(ctx, "key2")
	require.NoError(t, err)
	assert.True(t, allowed2)
}

func TestDefaultRateLimiterConfig(t *testing.T) {
	config := DefaultRateLimiterConfig()

	assert.Equal(t, 5, config.MaxAttempts)
	assert.Equal(t, time.Minute*15, config.Window)
	assert.Equal(t, time.Minute*30, config.LockoutTime)
}

func TestLoginRateLimiterConfig(t *testing.T) {
	config := LoginRateLimiterConfig()

	assert.Equal(t, 5, config.MaxAttempts)
	assert.Equal(t, time.Minute*15, config.Window)
	assert.Equal(t, time.Minute*30, config.LockoutTime)
}

func TestRegistrationRateLimiterConfig(t *testing.T) {
	config := RegistrationRateLimiterConfig()

	assert.Equal(t, 3, config.MaxAttempts)
	assert.Equal(t, time.Hour, config.Window)
	assert.Equal(t, time.Hour*2, config.LockoutTime)
}

func TestInMemoryRateLimiter_ConcurrentAccess(t *testing.T) {
	config := RateLimiterConfig{
		MaxAttempts: 10,
		Window:      time.Minute,
		LockoutTime: time.Minute,
	}
	limiter := NewInMemoryRateLimiter(config)
	ctx := context.Background()

	// Test concurrent access to the same key
	const numGoroutines = 5
	const attemptsPerGoroutine = 2

	results := make(chan bool, numGoroutines*attemptsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < attemptsPerGoroutine; j++ {
				allowed, err := limiter.Allow(ctx, "concurrent-key")
				if err == nil {
					results <- allowed
				} else {
					results <- false
				}
			}
		}()
	}

	// Collect results
	allowedCount := 0
	for i := 0; i < numGoroutines*attemptsPerGoroutine; i++ {
		if <-results {
			allowedCount++
		}
	}

	// All attempts should be allowed since we're within the limit
	assert.Equal(t, numGoroutines*attemptsPerGoroutine, allowedCount)

	// Final attempt count should be correct
	attempts, err := limiter.GetAttempts(ctx, "concurrent-key")
	require.NoError(t, err)
	assert.Equal(t, numGoroutines*attemptsPerGoroutine, attempts)
}
