package parcelemais

import (
	"net/url"
	"strings"
)

// ClientOptions configura o ParceleMaisClient.
type ClientOptions struct {
	ClientID     string
	ClientSecret string
	Environment  Environment        // ignorado se BaseURL for informado
	BaseURL      string             // opcional — sobrescreve Environment
	Resilience   *ResilienceOptions // opcional — nil usa DefaultResilienceOptions()
}

type resolvedClientOptions struct {
	clientID     string
	clientSecret string
	baseURL      string
	resilience   ResilienceOptions
}

func resolveClientOptions(opts ClientOptions) (*resolvedClientOptions, error) {
	if strings.TrimSpace(opts.ClientID) == "" {
		return nil, &ConfigurationError{Message: "ClientID é obrigatório."}
	}
	if strings.TrimSpace(opts.ClientSecret) == "" {
		return nil, &ConfigurationError{Message: "ClientSecret é obrigatório."}
	}

	baseURL := opts.BaseURL
	if baseURL == "" {
		env := opts.Environment
		if env == "" {
			env = EnvironmentProduction
		}
		mapped, ok := environmentBaseURLs[env]
		if !ok {
			return nil, &ConfigurationError{Message: "Environment inválido."}
		}
		baseURL = mapped
	} else {
		parsed, err := url.Parse(baseURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return nil, &ConfigurationError{Message: "BaseURL, quando informada, deve ser uma URL absoluta válida."}
		}
	}
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	resilience := DefaultResilienceOptions()
	if opts.Resilience != nil {
		resilience = *opts.Resilience
	}

	if err := validateResilience(resilience); err != nil {
		return nil, err
	}

	return &resolvedClientOptions{
		clientID:     opts.ClientID,
		clientSecret: opts.ClientSecret,
		baseURL:      baseURL,
		resilience:   resilience,
	}, nil
}

func validateResilience(r ResilienceOptions) error {
	if r.MaxRetryAttempts < 1 {
		return &ConfigurationError{Message: "Resilience.MaxRetryAttempts deve ser maior ou igual a 1."}
	}
	if r.TotalTimeout <= 0 {
		return &ConfigurationError{Message: "Resilience.TotalTimeout deve ser maior que zero."}
	}
	if r.AttemptTimeout <= 0 {
		return &ConfigurationError{Message: "Resilience.AttemptTimeout deve ser maior que zero."}
	}
	if r.AttemptTimeout > r.TotalTimeout {
		return &ConfigurationError{Message: "Resilience.AttemptTimeout não pode ser maior que Resilience.TotalTimeout."}
	}
	if r.CircuitBreakerFailureRatio <= 0 || r.CircuitBreakerFailureRatio > 1 {
		return &ConfigurationError{Message: "Resilience.CircuitBreakerFailureRatio deve estar entre 0 (exclusivo) e 1 (inclusivo)."}
	}
	if r.CircuitBreakerMinThroughput < 2 {
		return &ConfigurationError{Message: "Resilience.CircuitBreakerMinThroughput deve ser maior ou igual a 2."}
	}
	return nil
}
