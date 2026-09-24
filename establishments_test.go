package parcelemais_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	parcelemais "github.com/Twila-Digital/twila-parcelemais-go-sdk"
)

const establishmentID = "establishment-1"

func establishmentWire() map[string]interface{} {
	return map[string]interface{}{
		"estabelecimentoId": establishmentID,
		"documento":         "12345678000199",
		"razaoSocial":       "Loja Centro LTDA",
		"nomeFantasia":      "Loja Centro",
		"ativa":             true,
		"modeloDesembolso":  1,
		"responsavel": map[string]interface{}{
			"nome":    "Maria Souza",
			"email":   "maria@loja.com.br",
			"celular": "+5511999998888",
		},
		"contaBancaria": map[string]interface{}{
			"banco":         "341",
			"agencia":       "1234",
			"digitoAgencia": "",
			"conta":         "56789",
			"digitoConta":   "0",
			"tipoConta":     1,
		},
		"endereco": map[string]interface{}{
			"rua":    "Rua Exemplo",
			"numero": "100",
			"bairro": "Centro",
			"cidade": "São Paulo",
			"estado": "SP",
			"cep":    "01310100",
			"pais":   "Brasil",
		},
	}
}

func newCreateEstablishmentRequest() parcelemais.CreateEstablishmentRequest {
	return parcelemais.CreateEstablishmentRequest{
		Document:          "12345678000199",
		LegalName:         "Loja Centro LTDA",
		TradeName:         "Loja Centro",
		DisbursementModel: parcelemais.DisbursementModelEstablishmentChain,
		Owner: parcelemais.EstablishmentOwner{
			Name:  "Maria Souza",
			Email: "maria@loja.com.br",
			Phone: "+5511999998888",
		},
		BankAccount: parcelemais.EstablishmentBankAccount{
			BankNumber:    "341",
			AgencyNumber:  "1234",
			AccountNumber: "56789",
			AccountDigit:  "0",
			AccountType:   parcelemais.BankAccountTypeCurrent,
		},
		Address: parcelemais.EstablishmentAddress{
			Street:   "Rua Exemplo",
			Number:   "100",
			District: "Centro",
			City:     "São Paulo",
			State:    "SP",
			ZipCode:  "01310100",
		},
	}
}

func decodeBody(t *testing.T, r *http.Request) map[string]interface{} {
	t.Helper()

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("erro lendo o corpo: %v", err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("corpo não é JSON válido: %v", err)
	}

	return body
}

func TestEstablishmentsCreateSendsWireBodyAndReturnsID(t *testing.T) {
	var body map[string]interface{}
	var method string

	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/establishment", func(w http.ResponseWriter, r *http.Request) {
			method = r.Method
			body = decodeBody(t, r)
			writeJSON(w, http.StatusOK, map[string]interface{}{"estabelecimentoId": establishmentID})
		})
	})

	result, err := client.Establishments.Create(context.Background(), newCreateEstablishmentRequest())
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.EstablishmentID != establishmentID {
		t.Fatalf("EstablishmentID incorreto: %q", result.EstablishmentID)
	}
	if method != http.MethodPost {
		t.Fatalf("método incorreto: %s", method)
	}
	if body["documento"] != "12345678000199" {
		t.Fatalf("documento incorreto: %v", body["documento"])
	}
	if body["modeloDesembolso"] != float64(1) {
		t.Fatalf("modeloDesembolso incorreto: %v", body["modeloDesembolso"])
	}

	owner, ok := body["responsavel"].(map[string]interface{})
	if !ok || owner["celular"] != "+5511999998888" {
		t.Fatalf("responsavel incorreto: %v", body["responsavel"])
	}
	address, ok := body["endereco"].(map[string]interface{})
	if !ok {
		t.Fatalf("endereco deveria ter sido enviado: %v", body["endereco"])
	}
	if address["rua"] != "Rua Exemplo" || address["numero"] != "100" || address["bairro"] != "Centro" ||
		address["cidade"] != "São Paulo" || address["estado"] != "SP" || address["cep"] != "01310100" {
		t.Fatalf("endereco incorreto: %v", address)
	}
	if _, present := address["complemento"]; present {
		t.Fatalf("complemento vazio não deveria ser enviado: %v", address["complemento"])
	}
	if _, present := address["pais"]; present {
		t.Fatalf("pais vazio não deveria ser enviado: %v", address["pais"])
	}
}

func TestEstablishmentsGetMapsEstablishment(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/establishment/"+establishmentID, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, establishmentWire())
		})
	})

	establishment, err := client.Establishments.Get(context.Background(), establishmentID)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if establishment.TradeName != "Loja Centro" {
		t.Fatalf("TradeName incorreto: %q", establishment.TradeName)
	}
	if !establishment.IsActive {
		t.Fatal("IsActive deveria ser true")
	}
	if establishment.DisbursementModel == nil || *establishment.DisbursementModel != parcelemais.DisbursementModelEstablishmentChain {
		t.Fatalf("DisbursementModel incorreto: %v", establishment.DisbursementModel)
	}
	if establishment.BankAccount == nil || establishment.BankAccount.BankNumber != "341" {
		t.Fatalf("BankAccount incorreto: %+v", establishment.BankAccount)
	}
	if establishment.Address == nil || establishment.Address.City != "São Paulo" {
		t.Fatalf("Address incorreto: %+v", establishment.Address)
	}
}

func TestEstablishmentsGetWithoutBankAccountAndAddress(t *testing.T) {
	wire := establishmentWire()
	wire["ativa"] = false
	wire["modeloDesembolso"] = nil
	wire["contaBancaria"] = nil
	wire["endereco"] = nil

	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/establishment/"+establishmentID, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, wire)
		})
	})

	establishment, err := client.Establishments.Get(context.Background(), establishmentID)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if establishment.IsActive {
		t.Fatal("IsActive deveria ser false")
	}
	if establishment.DisbursementModel != nil {
		t.Fatalf("DisbursementModel deveria ser nil: %v", establishment.DisbursementModel)
	}
	if establishment.BankAccount != nil || establishment.Address != nil {
		t.Fatal("BankAccount e Address deveriam ser nil")
	}
}

func TestEstablishmentsListBuildsQueryString(t *testing.T) {
	var query string

	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/establishment/list", func(w http.ResponseWriter, r *http.Request) {
			query = r.URL.RawQuery
			writeJSON(w, http.StatusOK, []interface{}{establishmentWire()})
		})
	})

	isActive := true
	establishments, err := client.Establishments.List(context.Background(), parcelemais.ListEstablishmentsRequest{
		TradeName: "Centro",
		IsActive:  &isActive,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(establishments) != 1 || establishments[0].TradeName != "Loja Centro" {
		t.Fatalf("lojas incorretas: %+v", establishments)
	}
	if query != "ativa=true&nomeFantasia=Centro" {
		t.Fatalf("query string incorreta: %q", query)
	}
}

func TestEstablishmentsListWithoutFiltersSendsNoQuery(t *testing.T) {
	var query string

	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/establishment/list", func(w http.ResponseWriter, r *http.Request) {
			query = r.URL.RawQuery
			writeJSON(w, http.StatusOK, []interface{}{})
		})
	})

	establishments, err := client.Establishments.List(context.Background(), parcelemais.ListEstablishmentsRequest{})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(establishments) != 0 {
		t.Fatalf("esperava lista vazia: %+v", establishments)
	}
	if query != "" {
		t.Fatalf("não deveria enviar query string: %q", query)
	}
}

func TestEstablishmentsListInactiveSendsFalse(t *testing.T) {
	var query string

	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/establishment/list", func(w http.ResponseWriter, r *http.Request) {
			query = r.URL.RawQuery
			writeJSON(w, http.StatusOK, []interface{}{})
		})
	})

	isActive := false
	if _, err := client.Establishments.List(context.Background(), parcelemais.ListEstablishmentsRequest{IsActive: &isActive}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if query != "ativa=false" {
		t.Fatalf("query string incorreta: %q", query)
	}
}

func TestEstablishmentsUpdateSendsOnlyEditableFields(t *testing.T) {
	var body map[string]interface{}
	var method string

	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/establishment/"+establishmentID, func(w http.ResponseWriter, r *http.Request) {
			method = r.Method
			body = decodeBody(t, r)
			w.WriteHeader(http.StatusOK)
		})
	})

	err := client.Establishments.Update(context.Background(), establishmentID, parcelemais.UpdateEstablishmentRequest{
		TradeName: "Loja Centro Matriz",
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if method != http.MethodPut {
		t.Fatalf("método incorreto: %s", method)
	}
	if body["nomeFantasia"] != "Loja Centro Matriz" {
		t.Fatalf("nomeFantasia incorreto: %v", body["nomeFantasia"])
	}
	if _, present := body["contaBancaria"]; present {
		t.Fatal("contaBancaria não pertence mais ao update")
	}
	if _, present := body["modeloDesembolso"]; present {
		t.Fatalf("modeloDesembolso nulo deveria ser omitido: %v", body["modeloDesembolso"])
	}
}

func TestEstablishmentsUpdateBankAccountUsesOwnEndpoint(t *testing.T) {
	var body map[string]interface{}

	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/establishment/"+establishmentID+"/bank-account", func(w http.ResponseWriter, r *http.Request) {
			body = decodeBody(t, r)
			w.WriteHeader(http.StatusOK)
		})
	})

	err := client.Establishments.UpdateBankAccount(context.Background(), establishmentID, parcelemais.EstablishmentBankAccount{
		BankNumber:    "237",
		AgencyNumber:  "4321",
		AccountNumber: "98765",
		AccountDigit:  "1",
		AccountType:   parcelemais.BankAccountTypeSavings,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if body["banco"] != "237" {
		t.Fatalf("banco incorreto: %v", body["banco"])
	}
	if body["tipoConta"] != float64(2) {
		t.Fatalf("tipoConta incorreto: %v", body["tipoConta"])
	}
}

func TestEstablishmentsActivateAndDeactivateSendStatus(t *testing.T) {
	var bodies []map[string]interface{}

	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/establishment/"+establishmentID+"/status", func(w http.ResponseWriter, r *http.Request) {
			bodies = append(bodies, decodeBody(t, r))
			w.WriteHeader(http.StatusOK)
		})
	})

	if err := client.Establishments.Deactivate(context.Background(), establishmentID); err != nil {
		t.Fatalf("erro inesperado ao inativar: %v", err)
	}
	if err := client.Establishments.Activate(context.Background(), establishmentID); err != nil {
		t.Fatalf("erro inesperado ao ativar: %v", err)
	}

	if len(bodies) != 2 {
		t.Fatalf("esperava 2 chamadas, veio %d", len(bodies))
	}
	if bodies[0]["ativa"] != false {
		t.Fatalf("deactivate deveria enviar ativa=false: %v", bodies[0]["ativa"])
	}
	if bodies[1]["ativa"] != true {
		t.Fatalf("activate deveria enviar ativa=true: %v", bodies[1]["ativa"])
	}
}

func TestEstablishmentsCreateConflictMapsToAPIError(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/establishment", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusConflict, map[string]interface{}{"detalhe": "Documento já cadastrado."})
		})
	})

	_, err := client.Establishments.Create(context.Background(), newCreateEstablishmentRequest())
	if err == nil {
		t.Fatal("esperava erro")
	}

	var apiErr *parcelemais.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("esperava *APIError, veio %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusConflict {
		t.Fatalf("StatusCode incorreto: %d", apiErr.StatusCode)
	}
}

func TestEstablishmentsGetNotFoundMapsToAPIError(t *testing.T) {
	client := newTestClient(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/v1/establishment/missing", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{"detalhe": "Estabelecimento não encontrado."})
		})
	})

	_, err := client.Establishments.Get(context.Background(), "missing")
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
