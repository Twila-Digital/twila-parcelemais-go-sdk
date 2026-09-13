package parcelemais_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	parcelemais "github.com/Twila-Digital/twila-parcelemais-go-sdk"
)

func TestAuthRefreshesTokenAndRetriesOnceOn401(t *testing.T) {
	tokenCalls := 0
	var seenAuthHeaders []string

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/authentication/accesstoken", func(w http.ResponseWriter, r *http.Request) {
		tokenCalls++
		token := "stale-token"
		if tokenCalls > 1 {
			token = "fresh-token"
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"token_de_acesso": token, "expira_em_segundos": 3600, "tipo_de_token": "Bearer",
		})
	})
	mux.HandleFunc("/v1/order/order-1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		seenAuthHeaders = append(seenAuthHeaders, auth)
		if auth == "Bearer stale-token" {
			writeJSON(w, http.StatusUnauthorized, map[string]interface{}{"detalhe": "token expirado"})
			return
		}
		writeJSON(w, http.StatusOK, orderWire())
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := parcelemais.NewClient(parcelemais.ClientOptions{
		ClientID: "cid", ClientSecret: "csecret", BaseURL: server.URL + "/",
	})
	if err != nil {
		t.Fatalf("NewClient falhou: %v", err)
	}

	order, err := client.Orders.Get(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if order.ID != "order-1" {
		t.Fatalf("ID incorreto: %q", order.ID)
	}

	expected := []string{"Bearer stale-token", "Bearer fresh-token"}
	if len(seenAuthHeaders) != len(expected) {
		t.Fatalf("esperava %v, veio %v", expected, seenAuthHeaders)
	}
	for i, h := range expected {
		if seenAuthHeaders[i] != h {
			t.Fatalf("header %d: esperava %q, veio %q", i, h, seenAuthHeaders[i])
		}
	}
}
