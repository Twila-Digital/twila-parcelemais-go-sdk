package parcelemais_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestWebhooksListAuditMapsPagedResult(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/webhooks/auditoria", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("esperava GET, veio %s", r.Method)
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"itens": []interface{}{
					map[string]interface{}{
						"id":          "audit-1",
						"tipo":        3,
						"requisicao":  `{"id_pedido":"order-1"}`,
						"resposta":    "ok",
						"statusCode":  200,
						"dataCriacao": "2026-09-24T10:00:00Z",
					},
				},
				"pagina": map[string]interface{}{
					"tem_proximo": true, "tem_anterior": false, "numero": 1, "tamanho": 10, "total": 25,
				},
			})
		})
	})

	page, err := client.Webhooks.ListAudit(context.Background(), parcelemais.ListWebhookAuditRequest{})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items incorretos: %+v", page.Items)
	}
	item := page.Items[0]
	if item.ID != "audit-1" || item.Type != parcelemais.WebHookTypeOrder || item.StatusCode != 200 ||
		item.Request != `{"id_pedido":"order-1"}` || item.Response != "ok" || item.CreatedAt != "2026-09-24T10:00:00Z" {
		t.Fatalf("item incorreto: %+v", item)
	}
	if !page.HasNext || page.HasPrevious || page.PageNumber != 1 || page.PageSize != 10 || page.TotalCount != 25 {
		t.Fatalf("paginação incorreta: %+v", page)
	}
}

func TestWebhooksListAuditSendsOnlySetFiltersAndPageDefaults(t *testing.T) {
	var query url.Values
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/webhooks/auditoria", func(w http.ResponseWriter, r *http.Request) {
			query = r.URL.Query()
			writeJSON(w, http.StatusOK, map[string]interface{}{"itens": []interface{}{}, "pagina": map[string]interface{}{}})
		})
	})

	if _, err := client.Webhooks.ListAudit(context.Background(), parcelemais.ListWebhookAuditRequest{}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(query) != 2 || query.Get("pagina") != "1" || query.Get("tamanhoPagina") != "10" {
		t.Fatalf("query incorreta sem filtros: %v", query)
	}

	orderNumber := int64(42)
	statusCode := 500
	_, err := client.Webhooks.ListAudit(context.Background(), parcelemais.ListWebhookAuditRequest{
		StartDate:   "2026-09-01T00:00:00+03:00",
		EndDate:     "2026-09-30T23:59:59-03:00",
		OrderID:     "3fa85f64-5717-4562-b3fc-2c963f66afa6",
		OrderNumber: &orderNumber,
		StatusCode:  &statusCode,
		Page:        2,
		PageSize:    50,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	// Offset "+03:00" só volta intacto se o SDK codificar "+" como %2B (um "+" cru vira espaço no servidor).
	expected := map[string]string{
		"dataInicio":    "2026-09-01T00:00:00+03:00",
		"dataFim":       "2026-09-30T23:59:59-03:00",
		"pedidoId":      "3fa85f64-5717-4562-b3fc-2c963f66afa6",
		"numeroPedido":  "42",
		"statusCode":    "500",
		"pagina":        "2",
		"tamanhoPagina": "50",
	}
	if len(query) != len(expected) {
		t.Fatalf("query incorreta: %v", query)
	}
	for key, value := range expected {
		if query.Get(key) != value {
			t.Fatalf("%s incorreto: esperava %q, veio %q", key, value, query.Get(key))
		}
	}
}

func TestWebhooksListAuditErrorMapsToAPIError(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/webhooks/auditoria", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{"detalhe": "statusCode inválido"})
		})
	})

	_, err := client.Webhooks.ListAudit(context.Background(), parcelemais.ListWebhookAuditRequest{})
	var apiErr *parcelemais.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("esperava *APIError, veio %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("StatusCode incorreto: %d", apiErr.StatusCode)
	}
}

func TestWebhooksListAuditMalformedResponseReturnsError(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/webhooks/auditoria", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{isto não é json"))
		})
	})

	_, err := client.Webhooks.ListAudit(context.Background(), parcelemais.ListWebhookAuditRequest{})
	if err == nil {
		t.Fatal("esperava erro ao desserializar resposta malformada, veio nil")
	}
}

func TestWebhooksListAuditNetworkErrorPropagates(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/authentication/accesstoken", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"token_de_acesso":    "test-token",
			"expira_em_segundos": 3600,
			"tipo_de_token":      "Bearer",
		})
	})
	mux.HandleFunc("/v1/webhooks", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []interface{}{})
	})
	server := httptest.NewServer(mux)

	resilience := fastResilienceOptions()
	client, err := parcelemais.NewClient(parcelemais.ClientOptions{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		BaseURL:      server.URL + "/",
		Resilience:   &resilience,
	})
	if err != nil {
		t.Fatalf("NewClient falhou: %v", err)
	}

	// Aquece o cache de token com uma chamada bem-sucedida, pra que o erro de rede a
	// seguir aconteça na chamada ao recurso (auditoria), não na geração do token.
	if _, err := client.Webhooks.List(context.Background()); err != nil {
		t.Fatalf("aquecimento do token falhou: %v", err)
	}

	server.Close()

	if _, err := client.Webhooks.ListAudit(context.Background(), parcelemais.ListWebhookAuditRequest{}); err == nil {
		t.Fatal("esperava erro de rede após fechar o servidor, veio nil")
	}
}
