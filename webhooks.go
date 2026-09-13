package parcelemais

import (
	"context"
	"fmt"

	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/transport"
	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/wire"
)

type WebHookType int

const (
	WebHookTypeCustomer   WebHookType = 1
	WebHookTypeSimulation WebHookType = 2
	WebHookTypeOrder      WebHookType = 3
)

type WebHookAuthenticationType int

const (
	WebHookAuthenticationTypeNone  WebHookAuthenticationType = 1
	WebHookAuthenticationTypeBasic WebHookAuthenticationType = 2
	WebHookAuthenticationTypeJWT   WebHookAuthenticationType = 3
)

type Webhook struct {
	Type               WebHookType
	URL                string
	AuthenticationType WebHookAuthenticationType
}

type CreateWebhookRequest struct {
	Type               WebHookType
	URL                string
	AuthenticationType WebHookAuthenticationType
	Credential         string
}

type CreateWebhookResult struct {
	SigningSecret string
}

type UpdateWebhookRequest struct {
	URL                string
	AuthenticationType WebHookAuthenticationType
	Credential         string
}

type OrderWebhookEvent struct {
	OrderID    string
	Status     OrderStatus
	StatusRaw  int
	StatusName string
}

type WebhooksClient struct {
	executor *transport.Executor
}

func (c *WebhooksClient) Create(ctx context.Context, req CreateWebhookRequest) (*CreateWebhookResult, error) {
	wireReq := wire.CreateWebhookRequest{
		Type:               int(req.Type),
		URL:                req.URL,
		AuthenticationType: int(req.AuthenticationType),
		Credential:         req.Credential,
	}

	raw, err := c.executor.Post(ctx, "v1/webhooks", wireReq, 0)
	if err != nil {
		return nil, translateExecutorError(err)
	}
	if apiErr := transport.EnsureSuccess(raw); apiErr != nil {
		return nil, translateExecutorError(apiErr)
	}

	var result wire.CreateWebhookResponse
	if err := unmarshalJSON(raw.Body, &result); err != nil {
		return nil, err
	}
	return &CreateWebhookResult{SigningSecret: result.SigningSecret}, nil
}

func (c *WebhooksClient) List(ctx context.Context) ([]Webhook, error) {
	raw, err := c.executor.Get(ctx, "v1/webhooks")
	if err != nil {
		return nil, translateExecutorError(err)
	}
	if apiErr := transport.EnsureSuccess(raw); apiErr != nil {
		return nil, translateExecutorError(apiErr)
	}

	var result []wire.Webhook
	if err := unmarshalJSON(raw.Body, &result); err != nil {
		return nil, err
	}

	webhooks := make([]Webhook, 0, len(result))
	for _, item := range result {
		webhooks = append(webhooks, Webhook{
			Type:               WebHookType(item.Type),
			URL:                item.URL,
			AuthenticationType: WebHookAuthenticationType(item.AuthenticationType),
		})
	}
	return webhooks, nil
}

func (c *WebhooksClient) Update(ctx context.Context, webhookType WebHookType, req UpdateWebhookRequest) error {
	wireReq := wire.UpdateWebhookRequest{
		URL:                req.URL,
		AuthenticationType: int(req.AuthenticationType),
		Credential:         req.Credential,
	}

	raw, err := c.executor.Put(ctx, fmt.Sprintf("v1/webhooks/%d", int(webhookType)), wireReq)
	if err != nil {
		return translateExecutorError(err)
	}
	return translateExecutorError(transport.EnsureSuccess(raw))
}

func (c *WebhooksClient) Delete(ctx context.Context, webhookType WebHookType) error {
	raw, err := c.executor.Delete(ctx, fmt.Sprintf("v1/webhooks/%d", int(webhookType)))
	if err != nil {
		return translateExecutorError(err)
	}
	return translateExecutorError(transport.EnsureSuccess(raw))
}
