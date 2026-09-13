package parcelemais

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/wire"
)

const replayTolerance = 5 * time.Minute

// ParseWebhookEvent decodifica o corpo bruto de um webhook de pedido. Se signatureHeader
// e signingSecret forem informados (não vazios), a assinatura HMAC-SHA256 e a janela de
// replay (5 minutos) são verificadas antes de decodificar o evento.
func ParseWebhookEvent(rawJSON []byte, signatureHeader string, signingSecret string) (*OrderWebhookEvent, error) {
	if signatureHeader != "" && signingSecret != "" {
		if err := verifySignature(rawJSON, signatureHeader, signingSecret); err != nil {
			return nil, err
		}
	}

	var w wire.OrderWebhookEvent
	if err := json.Unmarshal(rawJSON, &w); err != nil {
		return nil, &WebhookSignatureError{Message: "O corpo do webhook está vazio ou não é um JSON válido."}
	}

	return &OrderWebhookEvent{
		OrderID:    w.OrderID,
		Status:     orderStatusFromWireValue(w.StatusEnum),
		StatusRaw:  w.StatusEnum,
		StatusName: w.Status,
	}, nil
}

// ComputeWebhookSignature calcula a assinatura HMAC-SHA256 do payload, no mesmo formato
// usado pela API do Parcele+ ("{timestamp}.{payload}").
func ComputeWebhookSignature(signingSecret string, timestampSeconds int64, payload []byte) string {
	signedContent := fmt.Sprintf("%d.%s", timestampSeconds, payload)
	mac := hmac.New(sha256.New, []byte(signingSecret))
	mac.Write([]byte(signedContent))
	return hex.EncodeToString(mac.Sum(nil))
}

func verifySignature(rawJSON []byte, signatureHeader string, signingSecret string) error {
	timestamp, signature, err := parseSignatureHeader(signatureHeader)
	if err != nil {
		return err
	}

	computed := ComputeWebhookSignature(signingSecret, timestamp, rawJSON)
	if !hmac.Equal([]byte(computed), []byte(signature)) {
		return &WebhookSignatureError{Message: "A assinatura do webhook não confere."}
	}

	eventTime := time.Unix(timestamp, 0)
	elapsed := time.Since(eventTime)
	if elapsed < 0 {
		elapsed = -elapsed
	}
	if elapsed > replayTolerance {
		return &WebhookSignatureError{Message: "O timestamp do webhook está fora da janela de tolerância — possível replay."}
	}
	return nil
}

func parseSignatureHeader(signatureHeader string) (int64, string, error) {
	var timestamp int64
	var signature string
	haveTimestamp := false
	haveSignature := false

	for _, part := range strings.Split(signatureHeader, ",") {
		key, value, found := strings.Cut(part, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "t":
			if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
				timestamp = parsed
				haveTimestamp = true
			}
		case "v1":
			signature = strings.ToLower(value)
			haveSignature = true
		}
	}

	if !haveTimestamp || !haveSignature {
		return 0, "", &WebhookSignatureError{Message: fmt.Sprintf("Cabeçalho de assinatura malformado: '%s'.", signatureHeader)}
	}
	return timestamp, signature, nil
}
