package idempotency_test

import (
	"testing"

	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/idempotency"
)

func TestRequiresKey(t *testing.T) {
	cases := []struct {
		method, path string
		expected     bool
	}{
		{"POST", "v1/order", true},
		{"POST", "v1/order/start-cdc-sale", true},
		{"POST", "v1/order/invoice", true},
		{"POST", "v1/webhooks", true},
		{"POST", "v1/order/simulate-installments", false},
		{"GET", "v1/order", false},
	}

	for _, c := range cases {
		if got := idempotency.RequiresKey(c.method, c.path); got != c.expected {
			t.Errorf("RequiresKey(%q, %q) = %v, esperava %v", c.method, c.path, got, c.expected)
		}
	}
}

func TestIsRetrySafe(t *testing.T) {
	cases := []struct {
		method, path string
		hasKey       bool
		expected     bool
	}{
		{"GET", "v1/order/123", false, true},
		{"HEAD", "v1/order", false, true},
		{"PUT", "v1/webhooks/1", false, true},
		{"DELETE", "v1/webhooks/1", false, true},
		{"PUT", "v1/order/123", false, false},
		{"POST", "v1/order", true, true},
		{"POST", "v1/order", false, false},
		{"POST", "v1/order/simulate-values", false, false},
	}

	for _, c := range cases {
		if got := idempotency.IsRetrySafe(c.method, c.path, c.hasKey); got != c.expected {
			t.Errorf("IsRetrySafe(%q, %q, %v) = %v, esperava %v", c.method, c.path, c.hasKey, got, c.expected)
		}
	}
}
