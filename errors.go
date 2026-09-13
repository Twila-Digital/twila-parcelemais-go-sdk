package parcelemais

import "fmt"

// ProblemDetails é o corpo de erro padrão (RFC 7807-like) devolvido pela API do Parcele+.
type ProblemDetails struct {
	Type          string
	Title         string
	Status        int
	Detail        string
	Instance      string
	Errors        map[string][]string
	CorrelationID string
}

// ConfigurationError indica que ClientOptions está inválida (ex.: ClientID/ClientSecret ausentes).
type ConfigurationError struct {
	Message string
}

func (e *ConfigurationError) Error() string { return e.Message }

// AuthenticationError indica falha ao gerar/renovar o token de acesso.
type AuthenticationError struct {
	Message string
	Cause   error
}

func (e *AuthenticationError) Error() string { return e.Message }
func (e *AuthenticationError) Unwrap() error { return e.Cause }

// TimeoutError indica timeout de rede, timeout total do pipeline de resiliência, ou circuit breaker aberto.
type TimeoutError struct {
	Message string
	Cause   error
}

func (e *TimeoutError) Error() string { return e.Message }
func (e *TimeoutError) Unwrap() error { return e.Cause }

// WebhookSignatureError indica assinatura de webhook inválida, malformada, ou fora da janela de replay.
type WebhookSignatureError struct {
	Message string
}

func (e *WebhookSignatureError) Error() string { return e.Message }

// APIError é a base para qualquer erro de resposta da API (404, 409, 5xx, etc.).
type APIError struct {
	StatusCode int
	Details    ProblemDetails
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API do Parcele+ retornou %d: %s", e.StatusCode, e.message())
}

func (e *APIError) message() string {
	if e.Details.Detail != "" {
		return e.Details.Detail
	}
	if e.Details.Title != "" {
		return e.Details.Title
	}
	return fmt.Sprintf("status %d", e.StatusCode)
}

// ErrorCode retorna o campo "tipo" do ProblemDetails, quando presente.
func (e *APIError) ErrorCode() string { return e.Details.Type }

// FieldErrors retorna os erros de validação por campo, quando presentes.
func (e *APIError) FieldErrors() map[string][]string { return e.Details.Errors }

// CorrelationID retorna o id de correlação devolvido pela API, quando presente.
func (e *APIError) CorrelationID() string { return e.Details.CorrelationID }

// ValidationError é um APIError especificamente de status 400 com erros de campo.
type ValidationError struct {
	*APIError
}

// RateLimitError é um APIError especificamente de status 429.
type RateLimitError struct {
	*APIError
	RetryAfterMs int64
}
