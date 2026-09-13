package contracttests

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

const swaggerURL = "https://api.staging.parcelemais.com.br/integration/swagger/v1/swagger.json"

var expectedPaths = []string{
	"/v1/authentication/accesstoken",
	"/v1/order",
	"/v1/order/{id}",
	"/v1/order/paged",
	"/v1/order/start-cdc-sale",
	"/v1/order/invoice",
	"/v1/order/simulate-installments",
	"/v1/order/simulate-values",
	"/v1/customer/{id}",
	"/v1/customer/paged",
	"/v1/webhooks",
	"/v1/webhooks/{type}",
}

var pathParamPattern = regexp.MustCompile(`\{[^}]+}`)

var (
	schemaOnce sync.Once
	schema     map[string]interface{}
	schemaErr  error
)

func fetchStagingSchema(t *testing.T) map[string]interface{} {
	t.Helper()

	schemaOnce.Do(func() {
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(swaggerURL)
		if err != nil {
			schemaErr = err
			return
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			schemaErr = &unexpectedStatusError{status: resp.StatusCode}
			return
		}

		schemaErr = json.NewDecoder(resp.Body).Decode(&schema)
	})

	if schemaErr != nil {
		t.Fatalf("falha ao buscar o swagger.json de staging: %v", schemaErr)
	}
	return schema
}

type unexpectedStatusError struct{ status int }

func (e *unexpectedStatusError) Error() string {
	return "HTTP " + http.StatusText(e.status)
}

func normalize(path string) string {
	return pathParamPattern.ReplaceAllString(path, "")
}

func simpleTypeName(schemaKey string) string {
	if idx := strings.LastIndex(schemaKey, "."); idx >= 0 {
		return schemaKey[idx+1:]
	}
	return schemaKey
}

func TestEndpointExistsInStagingSchema(t *testing.T) {
	schema := fetchStagingSchema(t)
	paths, _ := schema["paths"].(map[string]interface{})

	for _, expectedPath := range expectedPaths {
		expectedPath := expectedPath
		t.Run(expectedPath, func(t *testing.T) {
			matches := false
			for realPath := range paths {
				if strings.HasSuffix(normalize(realPath), normalize(expectedPath)) {
					matches = true
					break
				}
			}
			if !matches {
				t.Errorf("endpoint %q não encontrado no swagger.json de staging — o SDK e o backend divergiram", expectedPath)
			}
		})
	}
}

func TestOrderResponseSchemaHasFieldsMapperExpects(t *testing.T) {
	schema := fetchStagingSchema(t)
	components, _ := schema["components"].(map[string]interface{})
	schemas, _ := components["schemas"].(map[string]interface{})

	var orderSchema map[string]interface{}
	for key, value := range schemas {
		if simpleTypeName(key) == "OrderIntegrationResponse" {
			orderSchema, _ = value.(map[string]interface{})
			break
		}
	}
	if orderSchema == nil {
		t.Fatal("não encontrei o schema OrderIntegrationResponse no swagger.json de staging")
	}

	properties, _ := orderSchema["properties"].(map[string]interface{})
	for _, field := range []string{"id", "numero", "status", "documentoCliente", "criadoEm"} {
		if _, ok := properties[field]; !ok {
			t.Errorf("campo %q esperado pelo mapper não existe (mais) no schema de staging", field)
		}
	}
}
