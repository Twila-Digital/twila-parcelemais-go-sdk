package parcelemais_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	parcelemais "github.com/Twila-Digital/twila-parcelemais-go-sdk"
)

func TestErrorMapping401ToAuthenticationError(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/order/order-1", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusUnauthorized, map[string]interface{}{"detalhe": "token inválido"})
		})
	})

	_, err := client.Orders.Get(context.Background(), "order-1")

	var authErr *parcelemais.AuthenticationError
	if !errors.As(err, &authErr) {
		t.Fatalf("esperava *AuthenticationError, veio %T: %v", err, err)
	}
}

func TestErrorMapping400WithFieldErrorsToValidationError(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/order/order-1", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"detalhe": "campos inválidos",
				"erros":   map[string][]string{"cpf": {"obrigatório"}},
			})
		})
	})

	_, err := client.Orders.Get(context.Background(), "order-1")

	var validationErr *parcelemais.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("esperava *ValidationError, veio %T: %v", err, err)
	}
	if validationErr.StatusCode != 400 {
		t.Fatalf("StatusCode incorreto: %d", validationErr.StatusCode)
	}
	fieldErrors := validationErr.FieldErrors()
	if len(fieldErrors["cpf"]) != 1 || fieldErrors["cpf"][0] != "obrigatório" {
		t.Fatalf("FieldErrors incorreto: %+v", fieldErrors)
	}
}

func TestErrorMapping400WithoutFieldErrorsToGenericAPIError(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/order/order-1", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{"detalhe": "bad request"})
		})
	})

	_, err := client.Orders.Get(context.Background(), "order-1")

	var validationErr *parcelemais.ValidationError
	if errors.As(err, &validationErr) {
		t.Fatal("não deveria mapear pra ValidationError sem erros de campo")
	}
	var apiErr *parcelemais.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("esperava *APIError, veio %T: %v", err, err)
	}
}

func TestErrorMapping429ToRateLimitError(t *testing.T) {
	client := newTestClientNoRetry(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/order/order-1", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "2")
			writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{"detalhe": "muitas requisições"})
		})
	})

	_, err := client.Orders.Get(context.Background(), "order-1")

	var rateLimitErr *parcelemais.RateLimitError
	if !errors.As(err, &rateLimitErr) {
		t.Fatalf("esperava *RateLimitError, veio %T: %v", err, err)
	}
	if rateLimitErr.RetryAfterMs != 2000 {
		t.Fatalf("RetryAfterMs incorreto: %d", rateLimitErr.RetryAfterMs)
	}
}

func TestErrorMappingErrorCodeReadsProblemDetailsType(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/order/order-1", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{
				"tipo": "pedido-nao-encontrado", "detalhe": "não encontrado",
			})
		})
	})

	_, err := client.Orders.Get(context.Background(), "order-1")

	var apiErr *parcelemais.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("esperava *APIError, veio %T: %v", err, err)
	}
	if apiErr.ErrorCode() != "pedido-nao-encontrado" {
		t.Fatalf("ErrorCode incorreto: %q", apiErr.ErrorCode())
	}
}
