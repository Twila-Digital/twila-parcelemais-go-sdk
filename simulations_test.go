package parcelemais_test

import (
	"context"
	"net/http"
	"testing"

	parcelemais "github.com/Twila-Digital/twila-parcelemais-go-sdk"
)

func TestSimulateInstallmentsMapsResponse(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/order/simulate-installments", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, []map[string]interface{}{
				{"valorTotalDebito": 1600.0, "prazo": 3, "valorParcela": 533.33},
			})
		})
	})

	installments, err := client.Simulations.SimulateInstallments(context.Background(), parcelemais.SimulateInstallmentsRequest{RequestedAmount: 1500.0})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(installments) != 1 || installments[0].Term != 3 || installments[0].InstallmentAmount != 533.33 {
		t.Fatalf("resultado incorreto: %+v", installments)
	}
}

func TestSimulateValuesMapsResponse(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/order/simulate-values", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"valoresEstabelecimento": map[string]interface{}{"valorVenda": 1500.0, "valorDesembolso": 1450.0},
				"valoresCliente":         map[string]interface{}{"valorParcela": 533.33},
			})
		})
	})

	result, err := client.Simulations.SimulateValues(context.Background(), parcelemais.SimulateValuesRequest{Amount: 1500.0, Term: 3})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.SaleAmount != 1500.0 || result.DisbursementAmount != 1450.0 || result.InstallmentAmount != 533.33 {
		t.Fatalf("resultado incorreto: %+v", result)
	}
}
