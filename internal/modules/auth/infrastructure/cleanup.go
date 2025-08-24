package infrastructure

import (
	"context"
	"log"
	"sync"
	"time"
)

// SessionCleanupService handles periodic cleanup of expired sessions
type SessionCleanupService struct {
	repo      SessionRepository
	interval  time.Duration
	stopCh    chan struct{}
	doneCh    chan struct{}
	closeOnce sync.Once
}

// NewSessionCleanupService creates a new session cleanup service
func NewSessionCleanupService(repo SessionRepository, interval time.Duration) *SessionCleanupService {
	return &SessionCleanupService{
		repo:     repo,
		interval: interval,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Start begins the periodic cleanup process
func (s *SessionCleanupService) Start(ctx context.Context) {
	go s.run(ctx)
}

// Stop stops the cleanup service
func (s *SessionCleanupService) Stop() {
	close(s.stopCh)
	<-s.doneCh
}

// run executes the cleanup loop
func (s *SessionCleanupService) run(ctx context.Context) {
	defer s.closeOnce.Do(func() {
		close(s.doneCh)
	})

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run initial cleanup
	s.cleanup(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("Session cleanup service stopped due to context cancellation")
			return
		case <-s.stopCh:
			log.Println("Session cleanup service stopped")
			return
		case <-ticker.C:
			s.cleanup(ctx)
		}
	}
}

// cleanup performs the actual cleanup operation
func (s *SessionCleanupService) cleanup(ctx context.Context) {
	deletedCount, err := s.repo.CleanupExpired(ctx)
	if err != nil {
		log.Printf("Error during session cleanup: %v", err)
		return
	}

	if deletedCount > 0 {
		log.Printf("Cleaned up %d expired sessions", deletedCount)
	}
}

// CleanupNow performs an immediate cleanup operation
func (s *SessionCleanupService) CleanupNow(ctx context.Context) (int64, error) {
	return s.repo.CleanupExpired(ctx)
}
