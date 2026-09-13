package parcelemais_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	parcelemais "github.com/Twila-Digital/twila-parcelemais-go-sdk"
)

const signingSecret = "whsec_test"

func rawEvent(orderID string, status int, statusName string) []byte {
	return []byte(fmt.Sprintf(`{"id_pedido":%q,"enum_status":%d,"status":%q}`, orderID, status, statusName))
}

func TestParseWebhookEventWithoutSignatureMapsFields(t *testing.T) {
	event, err := parcelemais.ParseWebhookEvent(rawEvent("order-1", 9, "Comprado"), "", "")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if event.OrderID != "order-1" || event.Status != parcelemais.OrderStatusPurchased || event.StatusRaw != 9 {
		t.Fatalf("evento incorreto: %+v", event)
	}
}

func TestParseWebhookEventInvalidJSONRaisesError(t *testing.T) {
	_, err := parcelemais.ParseWebhookEvent([]byte("not json"), "", "")
	assertWebhookSignatureError(t, err)
}

func TestParseWebhookEventValidSignaturePasses(t *testing.T) {
	payload := rawEvent("order-1", 9, "Comprado")
	timestamp := time.Now().Unix()
	signature := parcelemais.ComputeWebhookSignature(signingSecret, timestamp, payload)
	header := fmt.Sprintf("t=%d,v1=%s", timestamp, signature)

	event, err := parcelemais.ParseWebhookEvent(payload, header, signingSecret)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if event.OrderID != "order-1" {
		t.Fatalf("OrderID incorreto: %q", event.OrderID)
	}
}

func TestParseWebhookEventInvalidSignatureRaisesError(t *testing.T) {
	payload := rawEvent("order-1", 9, "Comprado")
	timestamp := time.Now().Unix()
	header := fmt.Sprintf("t=%d,v1=%s", timestamp, strings.Repeat("0", 64))

	_, err := parcelemais.ParseWebhookEvent(payload, header, signingSecret)
	assertWebhookSignatureError(t, err)
}

func TestParseWebhookEventExpiredTimestampRaisesReplayError(t *testing.T) {
	payload := rawEvent("order-1", 9, "Comprado")
	oldTimestamp := time.Now().Add(-6 * time.Minute).Unix()
	signature := parcelemais.ComputeWebhookSignature(signingSecret, oldTimestamp, payload)
	header := fmt.Sprintf("t=%d,v1=%s", oldTimestamp, signature)

	_, err := parcelemais.ParseWebhookEvent(payload, header, signingSecret)
	assertWebhookSignatureError(t, err)
}

func TestParseWebhookEventMalformedHeaderRaisesError(t *testing.T) {
	_, err := parcelemais.ParseWebhookEvent(rawEvent("order-1", 9, "Comprado"), "garbage-header", signingSecret)
	assertWebhookSignatureError(t, err)
}

func assertWebhookSignatureError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("esperava erro, veio nil")
	}
	var sigErr *parcelemais.WebhookSignatureError
	if !errors.As(err, &sigErr) {
		t.Fatalf("esperava *WebhookSignatureError, veio %T: %v", err, err)
	}
}
