package parcelemais

import "time"

// ResilienceOptions controla retry, circuit breaker e timeouts. Os campos usam time.Duration
// (idiomático em Go), diferente dos outros SDKs desta família que usam inteiros em milissegundos.
type ResilienceOptions struct {
	TotalTimeout                   time.Duration
	AttemptTimeout                 time.Duration
	InvoiceUploadAttemptTimeout    time.Duration
	MaxRetryAttempts               int
	RetryBaseDelay                 time.Duration
	CircuitBreakerFailureRatio     float64
	CircuitBreakerSamplingWindow   time.Duration
	CircuitBreakerMinThroughput    int
	CircuitBreakerBreakDuration    time.Duration
	RetryOn500                     bool
	DisableAutomaticIdempotencyKey bool
}

// DefaultResilienceOptions retorna os valores padrão, iguais (em espírito) aos outros SDKs.
func DefaultResilienceOptions() ResilienceOptions {
	return ResilienceOptions{
		TotalTimeout:                   30 * time.Second,
		AttemptTimeout:                 10 * time.Second,
		InvoiceUploadAttemptTimeout:    60 * time.Second,
		MaxRetryAttempts:               3,
		RetryBaseDelay:                 500 * time.Millisecond,
		CircuitBreakerFailureRatio:     0.5,
		CircuitBreakerSamplingWindow:   30 * time.Second,
		CircuitBreakerMinThroughput:    10,
		CircuitBreakerBreakDuration:    15 * time.Second,
		RetryOn500:                     false,
		DisableAutomaticIdempotencyKey: false,
	}
}
