package parcelemais_test

import (
	"errors"
	"testing"
	"time"

	parcelemais "github.com/Twila-Digital/twila-parcelemais-go-sdk"
)

func TestNewClientRequiresClientID(t *testing.T) {
	_, err := parcelemais.NewClient(parcelemais.ClientOptions{ClientSecret: "secret"})
	assertConfigurationError(t, err)
}

func TestNewClientRequiresClientSecret(t *testing.T) {
	_, err := parcelemais.NewClient(parcelemais.ClientOptions{ClientID: "id"})
	assertConfigurationError(t, err)
}

func TestNewClientRejectsInvalidBaseURL(t *testing.T) {
	_, err := parcelemais.NewClient(parcelemais.ClientOptions{ClientID: "id", ClientSecret: "secret", BaseURL: "not-a-url"})
	assertConfigurationError(t, err)
}

func TestNewClientSucceedsWithDefaults(t *testing.T) {
	client, err := parcelemais.NewClient(parcelemais.ClientOptions{ClientID: "id", ClientSecret: "secret"})
	if err != nil {
		t.Fatalf("esperava sucesso, veio erro: %v", err)
	}
	if client.Orders == nil || client.Simulations == nil || client.Customers == nil || client.Webhooks == nil {
		t.Fatal("todos os clients de recurso deveriam estar inicializados")
	}
}

func TestNewClientRejectsInvalidResilienceOptions(t *testing.T) {
	cases := map[string]parcelemais.ResilienceOptions{
		"zero retries":       {MaxRetryAttempts: 0, TotalTimeout: time.Second, AttemptTimeout: time.Second, CircuitBreakerFailureRatio: 0.5, CircuitBreakerMinThroughput: 2},
		"zero total timeout": {MaxRetryAttempts: 1, TotalTimeout: 0, AttemptTimeout: time.Second, CircuitBreakerFailureRatio: 0.5, CircuitBreakerMinThroughput: 2},
		"attempt > total":    {MaxRetryAttempts: 1, TotalTimeout: time.Second, AttemptTimeout: 2 * time.Second, CircuitBreakerFailureRatio: 0.5, CircuitBreakerMinThroughput: 2},
		"ratio zero":         {MaxRetryAttempts: 1, TotalTimeout: time.Second, AttemptTimeout: time.Second, CircuitBreakerFailureRatio: 0, CircuitBreakerMinThroughput: 2},
		"ratio above 1":      {MaxRetryAttempts: 1, TotalTimeout: time.Second, AttemptTimeout: time.Second, CircuitBreakerFailureRatio: 1.5, CircuitBreakerMinThroughput: 2},
		"min throughput < 2": {MaxRetryAttempts: 1, TotalTimeout: time.Second, AttemptTimeout: time.Second, CircuitBreakerFailureRatio: 0.5, CircuitBreakerMinThroughput: 1},
	}

	for name, resilience := range cases {
		resilience := resilience
		t.Run(name, func(t *testing.T) {
			_, err := parcelemais.NewClient(parcelemais.ClientOptions{
				ClientID:     "id",
				ClientSecret: "secret",
				Resilience:   &resilience,
			})
			assertConfigurationError(t, err)
		})
	}
}

func assertConfigurationError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("esperava erro, veio nil")
	}
	var configErr *parcelemais.ConfigurationError
	if !errors.As(err, &configErr) {
		t.Fatalf("esperava *ConfigurationError, veio %T: %v", err, err)
	}
}
