package resilience_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/resilience"
)

func fastOptions() resilience.Options {
	return resilience.Options{
		TotalTimeout:                 5 * time.Second,
		MaxRetryAttempts:             3,
		RetryBaseDelay:               time.Millisecond,
		CircuitBreakerFailureRatio:   0.5,
		CircuitBreakerSamplingWindow: 30 * time.Second,
		CircuitBreakerMinThroughput:  10,
		CircuitBreakerBreakDuration:  60 * time.Second,
	}
}

type netError struct{}

func (netError) Error() string   { return "network error" }
func (netError) Timeout() bool   { return true }
func (netError) Temporary() bool { return true }

var _ net.Error = netError{}

func isNetworkError(err error) bool {
	var ne net.Error
	return errors.As(err, &ne)
}

func TestSuccessfulResponseDoesNotRetry(t *testing.T) {
	pipeline := resilience.NewPipeline(fastOptions())
	calls := 0

	resp, err := pipeline.Execute(context.Background(), true, isNetworkError, func(ctx context.Context) (resilience.Response, error) {
		calls++
		return resilience.Response{StatusCode: 200}, nil
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resp.StatusCode != 200 || calls != 1 {
		t.Fatalf("esperava 1 chamada e status 200, veio %d chamadas e status %d", calls, resp.StatusCode)
	}
}

func TestTransientResponseRetriesWhenRetrySafe(t *testing.T) {
	pipeline := resilience.NewPipeline(fastOptions())
	calls := 0

	resp, err := pipeline.Execute(context.Background(), true, isNetworkError, func(ctx context.Context) (resilience.Response, error) {
		calls++
		if calls < 3 {
			return resilience.Response{StatusCode: 503}, nil
		}
		return resilience.Response{StatusCode: 200}, nil
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resp.StatusCode != 200 || calls != 3 {
		t.Fatalf("esperava 3 chamadas e status 200, veio %d chamadas e status %d", calls, resp.StatusCode)
	}
}

func TestTransientResponseDoesNotRetryWhenNotRetrySafe(t *testing.T) {
	pipeline := resilience.NewPipeline(fastOptions())
	calls := 0

	resp, err := pipeline.Execute(context.Background(), false, isNetworkError, func(ctx context.Context) (resilience.Response, error) {
		calls++
		return resilience.Response{StatusCode: 503}, nil
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resp.StatusCode != 503 || calls != 1 {
		t.Fatalf("esperava 1 chamada e status 503, veio %d chamadas e status %d", calls, resp.StatusCode)
	}
}

func TestNetworkErrorRetriesAndEventuallyReturnsError(t *testing.T) {
	pipeline := resilience.NewPipeline(fastOptions())
	calls := 0

	_, err := pipeline.Execute(context.Background(), true, isNetworkError, func(ctx context.Context) (resilience.Response, error) {
		calls++
		return resilience.Response{}, netError{}
	})
	if err == nil {
		t.Fatal("esperava erro")
	}
	if calls != 3 {
		t.Fatalf("esperava 3 chamadas, veio %d", calls)
	}
}

func TestTotalTimeoutReturnsTimeoutExceededError(t *testing.T) {
	opts := resilience.Options{
		TotalTimeout:                 1 * time.Millisecond,
		MaxRetryAttempts:             5,
		RetryBaseDelay:               50 * time.Millisecond,
		CircuitBreakerFailureRatio:   0.5,
		CircuitBreakerSamplingWindow: 30 * time.Second,
		CircuitBreakerMinThroughput:  10,
		CircuitBreakerBreakDuration:  60 * time.Second,
	}
	pipeline := resilience.NewPipeline(opts)

	_, err := pipeline.Execute(context.Background(), true, isNetworkError, func(ctx context.Context) (resilience.Response, error) {
		return resilience.Response{StatusCode: 503}, nil
	})

	var timeoutErr *resilience.TimeoutExceededError
	if !errors.As(err, &timeoutErr) {
		t.Fatalf("esperava *TimeoutExceededError, veio %T: %v", err, err)
	}
}

func TestCircuitBreakerOpensAfterFailureThresholdAndBlocksNextCall(t *testing.T) {
	opts := resilience.Options{
		TotalTimeout:                 5 * time.Second,
		MaxRetryAttempts:             1,
		RetryBaseDelay:               time.Millisecond,
		CircuitBreakerFailureRatio:   0.5,
		CircuitBreakerSamplingWindow: 30 * time.Second,
		CircuitBreakerMinThroughput:  2,
		CircuitBreakerBreakDuration:  60 * time.Second,
	}
	pipeline := resilience.NewPipeline(opts)

	for i := 0; i < 2; i++ {
		_, _ = pipeline.Execute(context.Background(), false, isNetworkError, func(ctx context.Context) (resilience.Response, error) {
			return resilience.Response{StatusCode: 503}, nil
		})
	}

	_, err := pipeline.Execute(context.Background(), false, isNetworkError, func(ctx context.Context) (resilience.Response, error) {
		return resilience.Response{StatusCode: 200}, nil
	})

	var timeoutErr *resilience.TimeoutExceededError
	if !errors.As(err, &timeoutErr) || !timeoutErr.BrokenCircuit {
		t.Fatalf("esperava *TimeoutExceededError com BrokenCircuit, veio %T: %v", err, err)
	}
}
