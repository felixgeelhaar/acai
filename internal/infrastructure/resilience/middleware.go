// Package resilience provides a repository decorator that wraps calls
// with fault-tolerance patterns from the Fortify library.
// This implements the decorator pattern (DDD infrastructure concern)
// without modifying the underlying repository.
package resilience

import (
	"context"
	"time"

	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
	"github.com/felixgeelhaar/fortify/circuitbreaker"
	"github.com/felixgeelhaar/fortify/ratelimit"
	"github.com/felixgeelhaar/fortify/retry"
	"github.com/felixgeelhaar/fortify/timeout"
)

// Config defines the resilience configuration.
type Config struct {
	Timeout          time.Duration
	MaxRetries       int
	RetryDelay       time.Duration
	RetryMaxDelay    time.Duration
	FailureThreshold uint32
	SuccessThreshold uint32
	HalfOpenTimeout  time.Duration
	RateLimit        int
	RateBurst        int
	RateInterval     time.Duration
}

// DefaultConfig returns production-safe defaults.
func DefaultConfig() Config {
	return Config{
		Timeout:          30 * time.Second,
		MaxRetries:       3,
		RetryDelay:       500 * time.Millisecond,
		RetryMaxDelay:    10 * time.Second,
		FailureThreshold: 5,
		SuccessThreshold: 2,
		HalfOpenTimeout:  60 * time.Second,
		RateLimit:        100,
		RateBurst:        150,
		RateInterval:     time.Minute,
	}
}

// ResilientRepository decorates a domain.Repository with resilience patterns
// using the Fortify library: circuit breaker, retry, rate limiting, and timeout.
type ResilientRepository struct {
	inner domain.Repository
	tm    timeout.Timeout[any]
	cfg   Config
	cb    circuitbreaker.CircuitBreaker[any]
	rt    retry.Retry[any]
	rl    ratelimit.RateLimiter
}

// NewResilientRepository creates a resilient repository decorator.
func NewResilientRepository(inner domain.Repository, cfg Config) *ResilientRepository {
	cb := circuitbreaker.New[any](circuitbreaker.Config{
		MaxRequests: cfg.SuccessThreshold,
		Timeout:     cfg.HalfOpenTimeout,
		ReadyToTrip: func(counts circuitbreaker.Counts) bool {
			return counts.ConsecutiveFailures >= cfg.FailureThreshold
		},
	})

	rt := retry.New[any](retry.Config{
		MaxAttempts:   cfg.MaxRetries,
		InitialDelay:  cfg.RetryDelay,
		MaxDelay:      cfg.RetryMaxDelay,
		BackoffPolicy: retry.BackoffExponential,
		Jitter:        true,
	})

	tm := timeout.New[any](timeout.Config{
		DefaultTimeout: cfg.Timeout,
	})

	rl := ratelimit.New(&ratelimit.Config{
		Rate:     cfg.RateLimit,
		Burst:    cfg.RateBurst,
		Interval: cfg.RateInterval,
	})

	return &ResilientRepository{
		inner: inner,
		cfg:   cfg,
		cb:    cb,
		rt:    rt,
		tm:    tm,
		rl:    rl,
	}
}

// Close releases resources held by the resilient repository.
func (r *ResilientRepository) Close() error {
	return r.rl.Close()
}

// execute runs the given function through the full resilience stack:
// rate limit → timeout → circuit breaker → retry → operation
func (r *ResilientRepository) execute(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
	if err := r.rl.Wait(ctx, "granola-api"); err != nil {
		return nil, err
	}
	return r.tm.Execute(ctx, 0, func(ctx context.Context) (any, error) {
		return r.cb.Execute(ctx, func(ctx context.Context) (any, error) {
			return r.rt.Do(ctx, fn)
		})
	})
}

func (r *ResilientRepository) FindByID(ctx context.Context, id domain.MeetingID) (*domain.Meeting, error) {
	result, err := r.execute(ctx, func(ctx context.Context) (any, error) {
		return r.inner.FindByID(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return result.(*domain.Meeting), nil
}

func (r *ResilientRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Meeting, error) {
	result, err := r.execute(ctx, func(ctx context.Context) (any, error) {
		return r.inner.List(ctx, filter)
	})
	if err != nil {
		return nil, err
	}
	return result.([]*domain.Meeting), nil
}

func (r *ResilientRepository) GetTranscript(ctx context.Context, id domain.MeetingID) (*domain.Transcript, error) {
	result, err := r.execute(ctx, func(ctx context.Context) (any, error) {
		return r.inner.GetTranscript(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return result.(*domain.Transcript), nil
}

func (r *ResilientRepository) SearchTranscripts(ctx context.Context, query string, filter domain.ListFilter) ([]*domain.Meeting, error) {
	result, err := r.execute(ctx, func(ctx context.Context) (any, error) {
		return r.inner.SearchTranscripts(ctx, query, filter)
	})
	if err != nil {
		return nil, err
	}
	return result.([]*domain.Meeting), nil
}

func (r *ResilientRepository) GetActionItems(ctx context.Context, id domain.MeetingID) ([]*domain.ActionItem, error) {
	result, err := r.execute(ctx, func(ctx context.Context) (any, error) {
		return r.inner.GetActionItems(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return result.([]*domain.ActionItem), nil
}

func (r *ResilientRepository) Sync(ctx context.Context, since *time.Time) ([]domain.DomainEvent, error) {
	result, err := r.execute(ctx, func(ctx context.Context) (any, error) {
		return r.inner.Sync(ctx, since)
	})
	if err != nil {
		return nil, err
	}
	return result.([]domain.DomainEvent), nil
}
