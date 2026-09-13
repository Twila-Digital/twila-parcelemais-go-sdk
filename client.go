// Package parcelemais é o SDK oficial em Go para a API do Parcele+ (crédito direto
// ao consumidor e parcelamento no momento da compra).
package parcelemais

import (
	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/auth"
	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/resilience"
	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/transport"
)

// Client é o cliente principal do SDK. Uma instância é segura para uso concorrente
// por múltiplas goroutines — reaproveite-a como singleton na aplicação (não crie uma
// por requisição); ela mantém o cache do token de acesso e o estado do circuit breaker.
type Client struct {
	Orders      *OrdersClient
	Simulations *SimulationsClient
	Customers   *CustomersClient
	Webhooks    *WebhooksClient
}

// NewClient valida options e constrói um Client pronto para uso.
func NewClient(options ClientOptions) (*Client, error) {
	resolved, err := resolveClientOptions(options)
	if err != nil {
		return nil, err
	}

	tokenAPIClient := auth.NewTokenAPIClient(resolved.baseURL, resolved.resilience.AttemptTimeout)
	tokenProvider := auth.NewTokenProvider(tokenAPIClient, resolved.clientID, resolved.clientSecret)

	resilienceOpts := resilience.Options{
		TotalTimeout:                 resolved.resilience.TotalTimeout,
		MaxRetryAttempts:             resolved.resilience.MaxRetryAttempts,
		RetryBaseDelay:               resolved.resilience.RetryBaseDelay,
		CircuitBreakerFailureRatio:   resolved.resilience.CircuitBreakerFailureRatio,
		CircuitBreakerSamplingWindow: resolved.resilience.CircuitBreakerSamplingWindow,
		CircuitBreakerMinThroughput:  resolved.resilience.CircuitBreakerMinThroughput,
		CircuitBreakerBreakDuration:  resolved.resilience.CircuitBreakerBreakDuration,
		RetryOn500:                   resolved.resilience.RetryOn500,
	}

	executor := transport.NewExecutor(
		resolved.baseURL,
		tokenProvider,
		resilienceOpts,
		resolved.resilience.AttemptTimeout,
		resolved.resilience.DisableAutomaticIdempotencyKey,
	)

	return &Client{
		Orders:      &OrdersClient{executor: executor, invoiceUploadAttemptTimeout: resolved.resilience.InvoiceUploadAttemptTimeout},
		Simulations: &SimulationsClient{executor: executor},
		Customers:   &CustomersClient{executor: executor},
		Webhooks:    &WebhooksClient{executor: executor},
	}, nil
}
