package parcelemais_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	parcelemais "github.com/Twila-Digital/twila-parcelemais-go-sdk"
)

func orderWire() map[string]interface{} {
	return map[string]interface{}{
		"id":                         "order-1",
		"numero":                     42,
		"status":                     map[string]interface{}{"valor": 9, "descricao": "Comprado"},
		"documentoCliente":           "12345678901",
		"razaoSocialEstabelecimento": "Loja Exemplo",
		"documentoEstabelecimento":   "12345678000195",
		"criadoEm":                   "2026-01-01T00:00:00-03:00",
	}
}

func TestOrdersCreateReturnsOrderID(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/order", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]string{"pedidoId": "order-1"})
		})
	})

	orderID, err := client.Orders.Create(context.Background(), parcelemais.CreateOrderRequest{
		CPF:                   "12345678901",
		PhoneNumber:           "+5511999998888",
		EstablishmentDocument: "12345678000195",
		RequestedAmount:       1500.0,
		Name:                  "Maria Souza",
		Email:                 "maria@exemplo.com.br",
		DateOfBirth:           "1990-05-20T00:00:00-03:00",
		Address: parcelemais.Address{
			Street: "Av. Paulista", Number: "1578", Neighborhood: "Bela Vista",
			City: "São Paulo", State: "SP", PostalCode: "01311000",
		},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if orderID != "order-1" {
		t.Fatalf("esperava order-1, veio %q", orderID)
	}
}

func TestOrdersGetMapsStatusAndFields(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/order/order-1", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, orderWire())
		})
	})

	order, err := client.Orders.Get(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if order.Status != parcelemais.OrderStatusPurchased {
		t.Fatalf("esperava OrderStatusPurchased, veio %v", order.Status)
	}
	if order.StatusDescription != "Comprado" {
		t.Fatalf("StatusDescription incorreto: %q", order.StatusDescription)
	}
	if order.CustomerDocument != "12345678901" {
		t.Fatalf("CustomerDocument incorreto: %q", order.CustomerDocument)
	}
}

func TestOrdersListMapsPagedResult(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/order/paged", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"itens": []interface{}{orderWire()},
				"pagina": map[string]interface{}{
					"tem_proximo": true, "tem_anterior": false, "numero": 1, "tamanho": 10, "total": 25,
				},
			})
		})
	})

	page, err := client.Orders.List(context.Background(), parcelemais.ListOrdersRequest{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "order-1" {
		t.Fatalf("items incorretos: %+v", page.Items)
	}
	if !page.HasNext || page.TotalCount != 25 {
		t.Fatalf("paginação incorreta: %+v", page)
	}
}

func TestOrdersStartCdcSaleReturnsCheckoutLink(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/order/start-cdc-sale", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]string{"linkPagamento": "https://pay.example.com/abc"})
		})
	})

	link, err := client.Orders.StartCdcSale(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if link.URL != "https://pay.example.com/abc" {
		t.Fatalf("URL incorreta: %q", link.URL)
	}
}

func TestOrdersGetNotFoundMapsToAPIError(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/order/missing", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{"detalhe": "não encontrado"})
		})
	})

	_, err := client.Orders.Get(context.Background(), "missing")
	if err == nil {
		t.Fatal("esperava erro")
	}
	var apiErr *parcelemais.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("esperava *APIError, veio %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("StatusCode incorreto: %d", apiErr.StatusCode)
	}
}
