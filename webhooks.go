package parcelemais

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

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

// WebhookAudit é um registro de auditoria de entrega de webhook — um por tentativa de
// entrega. Request/Response trazem o corpo enviado e o corpo recebido do endpoint do
// integrador; CreatedAt vem como string ISO-8601, igual aos demais campos de data do SDK.
type WebhookAudit struct {
	ID         string
	Type       WebHookType
	Request    string
	Response   string
	StatusCode int
	CreatedAt  string
}

// ListWebhookAuditRequest filtra a auditoria de entregas. Todos os filtros são opcionais:
// strings vazias e ponteiros nil são omitidos da query. StartDate/EndDate são date-time
// ISO-8601; StatusCode é o status HTTP devolvido pelo endpoint do integrador (100–599).
type ListWebhookAuditRequest struct {
	StartDate   string
	EndDate     string
	OrderID     string
	OrderNumber *int64
	StatusCode  *int
	Page        int
	PageSize    int
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

// ListAudit lista a auditoria de entregas de webhook, paginada, da mais recente pra mais
// antiga (um registro por tentativa de entrega).
func (c *WebhooksClient) ListAudit(ctx context.Context, req ListWebhookAuditRequest) (*PagedResult[WebhookAudit], error) {
	q := url.Values{}
	if req.StartDate != "" {
		q.Set("dataInicio", req.StartDate)
	}
	if req.EndDate != "" {
		q.Set("dataFim", req.EndDate)
	}
	if req.OrderID != "" {
		q.Set("pedidoId", req.OrderID)
	}
	if req.OrderNumber != nil {
		q.Set("numeroPedido", strconv.FormatInt(*req.OrderNumber, 10))
	}
	if req.StatusCode != nil {
		q.Set("statusCode", strconv.Itoa(*req.StatusCode))
	}
	q.Set("pagina", strconv.Itoa(pageOrDefault(req.Page)))
	q.Set("tamanhoPagina", strconv.Itoa(pageSizeOrDefault(req.PageSize)))

	raw, err := c.executor.Get(ctx, "v1/webhooks/auditoria?"+q.Encode())
	if err != nil {
		return nil, translateExecutorError(err)
	}
	if apiErr := transport.EnsureSuccess(raw); apiErr != nil {
		return nil, translateExecutorError(apiErr)
	}

	var result wire.PagedWebhookAudits
	if err := unmarshalJSON(raw.Body, &result); err != nil {
		return nil, err
	}

	items := make([]WebhookAudit, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, WebhookAudit{
			ID:         item.ID,
			Type:       WebHookType(item.Type),
			Request:    item.Request,
			Response:   item.Response,
			StatusCode: item.StatusCode,
			CreatedAt:  item.CreatedAt,
		})
	}

	return &PagedResult[WebhookAudit]{
		Items:       items,
		HasNext:     result.Pagina.HasNext,
		HasPrevious: result.Pagina.HasPrevious,
		PageNumber:  result.Pagina.Number,
		PageSize:    result.Pagina.Size,
		TotalCount:  result.Pagina.Total,
	}, nil
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
