package parcelemais_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	parcelemais "github.com/Twila-Digital/twila-parcelemais-go-sdk"
)

// fastResilienceOptions evita esperar o backoff real (segundos) nos testes — o
// comportamento de retry/circuit breaker em si já tem cobertura dedicada em
// internal/resilience.
func fastResilienceOptions() parcelemais.ResilienceOptions {
	return parcelemais.ResilienceOptions{
		TotalTimeout:                   5 * time.Second,
		AttemptTimeout:                 2 * time.Second,
		InvoiceUploadAttemptTimeout:    5 * time.Second,
		MaxRetryAttempts:               3,
		RetryBaseDelay:                 time.Millisecond,
		CircuitBreakerFailureRatio:     0.5,
		CircuitBreakerSamplingWindow:   30 * time.Second,
		CircuitBreakerMinThroughput:    1000, // alto o bastante pra não abrir sozinho durante os testes
		CircuitBreakerBreakDuration:    time.Second,
		DisableAutomaticIdempotencyKey: false,
	}
}

func newTestServer(t *testing.T, register func(mux *http.ServeMux)) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/authentication/accesstoken", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"token_de_acesso":    "test-token",
			"expira_em_segundos": 3600,
			"tipo_de_token":      "Bearer",
		})
	})
	if register != nil {
		register(mux)
	}

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func newTestClient(t *testing.T, register func(mux *http.ServeMux)) *parcelemais.Client {
	t.Helper()
	resilience := fastResilienceOptions()
	return newTestClientWithResilience(t, register, resilience)
}

// newTestClientNoRetry usa MaxRetryAttempts: 1 — útil pra testar mapeamento de erro em
// status transientes (429, 503, etc.) sem esperar o Retry-After real do handler de teste.
func newTestClientNoRetry(t *testing.T, register func(mux *http.ServeMux)) *parcelemais.Client {
	t.Helper()
	resilience := fastResilienceOptions()
	resilience.MaxRetryAttempts = 1
	return newTestClientWithResilience(t, register, resilience)
}

func newTestClientWithResilience(t *testing.T, register func(mux *http.ServeMux), resilience parcelemais.ResilienceOptions) *parcelemais.Client {
	t.Helper()

	server := newTestServer(t, register)

	client, err := parcelemais.NewClient(parcelemais.ClientOptions{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		BaseURL:      server.URL + "/",
		Resilience:   &resilience,
	})
	if err != nil {
		t.Fatalf("NewClient falhou: %v", err)
	}
	return client
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
