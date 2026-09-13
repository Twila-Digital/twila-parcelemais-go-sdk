package parcelemais_test

import (
	"context"
	"net/http"
	"testing"

	parcelemais "github.com/Twila-Digital/twila-parcelemais-go-sdk"
)

func customerWire() map[string]interface{} {
	return map[string]interface{}{
		"id":               "customer-1",
		"nome":             "Maria Souza",
		"documento":        "12345678901",
		"dataDeNascimento": "1990-05-20T00:00:00-03:00",
		"endereco":         map[string]interface{}{"rua": "Av. Paulista", "cidade": "São Paulo", "estado": "SP"},
		"email":            "maria@exemplo.com.br",
	}
}

func TestCustomersGetMapsCustomer(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/customer/customer-1", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, customerWire())
		})
	})

	customer, err := client.Customers.Get(context.Background(), "customer-1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if customer.Name != "Maria Souza" {
		t.Fatalf("Name incorreto: %q", customer.Name)
	}
	if customer.Address == nil || *customer.Address.City != "São Paulo" {
		t.Fatalf("Address incorreto: %+v", customer.Address)
	}
}

func TestCustomersListMapsPagedResult(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/customer/paged", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"itens":  []interface{}{customerWire()},
				"pagina": map[string]interface{}{"tem_proximo": false, "tem_anterior": false, "numero": 1, "tamanho": 10, "total": 1},
			})
		})
	})

	page, err := client.Customers.List(context.Background(), parcelemais.ListCustomersRequest{})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "customer-1" {
		t.Fatalf("items incorretos: %+v", page.Items)
	}
	if page.HasNext {
		t.Fatal("HasNext deveria ser false")
	}
}
