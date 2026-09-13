package parcelemais_test

import (
	"context"
	"net/http"
	"testing"

	parcelemais "github.com/Twila-Digital/twila-parcelemais-go-sdk"
)

func TestWebhooksCreateReturnsSigningSecret(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/webhooks", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				writeJSON(w, http.StatusOK, map[string]string{"chaveAssinatura": "whsec_abc"})
				return
			}
			writeJSON(w, http.StatusOK, []interface{}{})
		})
	})

	result, err := client.Webhooks.Create(context.Background(), parcelemais.CreateWebhookRequest{
		Type:               parcelemais.WebHookTypeOrder,
		URL:                "https://example.com/webhook",
		AuthenticationType: parcelemais.WebHookAuthenticationTypeNone,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.SigningSecret != "whsec_abc" {
		t.Fatalf("SigningSecret incorreto: %q", result.SigningSecret)
	}
}

func TestWebhooksListMapsWebhooks(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/webhooks", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, []map[string]interface{}{
				{"tipo": 3, "url": "https://example.com/webhook", "tipoAutenticacao": 1},
			})
		})
	})

	webhooks, err := client.Webhooks.List(context.Background())
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(webhooks) != 1 || webhooks[0].Type != parcelemais.WebHookTypeOrder {
		t.Fatalf("webhooks incorretos: %+v", webhooks)
	}
}

func TestWebhooksUpdateSendsPut(t *testing.T) {
	called := false
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/webhooks/3", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Fatalf("esperava PUT, veio %s", r.Method)
			}
			called = true
			w.WriteHeader(http.StatusNoContent)
		})
	})

	err := client.Webhooks.Update(context.Background(), parcelemais.WebHookTypeOrder, parcelemais.UpdateWebhookRequest{
		URL: "https://example.com/new", AuthenticationType: parcelemais.WebHookAuthenticationTypeNone,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !called {
		t.Fatal("handler PUT não foi chamado")
	}
}

func TestWebhooksDeleteSendsDelete(t *testing.T) {
	called := false
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/webhooks/3", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Fatalf("esperava DELETE, veio %s", r.Method)
			}
			called = true
			w.WriteHeader(http.StatusNoContent)
		})
	})

	err := client.Webhooks.Delete(context.Background(), parcelemais.WebHookTypeOrder)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !called {
		t.Fatal("handler DELETE não foi chamado")
	}
}
