package parcelemais

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/transport"
	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/wire"
)

// OrderStatus representa o status de um pedido. Go não precisa de fallback pra "valor
// desconhecido" como os outros SDKs desta família — um int fora do conjunto conhecido
// já é representável e comparável normalmente; OrderStatusUnknown existe só por simetria
// com os outros SDKs (e cobre o caso raro de "valor" vir omitido/zero por engano).
type OrderStatus int

const (
	OrderStatusUndefined                  OrderStatus = 0
	OrderStatusAnalysing                  OrderStatus = 1
	OrderStatusApproved                   OrderStatus = 2
	OrderStatusUnavailableBalance         OrderStatus = 3
	OrderStatusAnalysisExpired            OrderStatus = 4
	OrderStatusPendingPayment             OrderStatus = 5
	OrderStatusBiometryRefused            OrderStatus = 6
	OrderStatusBiometryApproved           OrderStatus = 7
	OrderStatusPaymentRefused             OrderStatus = 8
	OrderStatusPurchased                  OrderStatus = 9
	OrderStatusUnauthorized               OrderStatus = 10
	OrderStatusPendingAuthorization       OrderStatus = 11
	OrderStatusAwaitingRegistration       OrderStatus = 12
	OrderStatusSaleNotStarted             OrderStatus = 13
	OrderStatusCanceled                   OrderStatus = 14
	OrderStatusBilling                    OrderStatus = 15
	OrderStatusCompleted                  OrderStatus = 16
	OrderStatusFrozen                     OrderStatus = 17
	OrderStatusPendingPaymentConfirmation OrderStatus = 18
	OrderStatusDisbursed                  OrderStatus = 19
	OrderStatusUnknown                    OrderStatus = -1
)

type Address struct {
	Street       string
	Number       string
	Neighborhood string
	City         string
	State        string
	PostalCode   string
	Complement   string
}

type CreateOrderRequest struct {
	CPF                   string
	PhoneNumber           string
	EstablishmentDocument string
	RequestedAmount       float64
	Name                  string
	Email                 string
	DateOfBirth           string
	Address               Address
}

type Order struct {
	ID                     string
	Number                 int64
	Status                 OrderStatus
	StatusDescription      string
	CustomerDocument       string
	EstablishmentLegalName string
	EstablishmentDocument  string
	CreatedAt              string
	Total                  *float64
	CustomerName           *string
	Term                   *int64
	Description            *string
	ApprovedAmount         *float64
	Disbursed              *bool
	DisbursedAt            *string
	RequestedAmount        *float64
}

type ListOrdersRequest struct {
	Status                *OrderStatus
	CustomerDocument      string
	StartDate             string
	EndDate               string
	Number                *int64
	EstablishmentDocument string
	Description           string
	Page                  int
	PageSize              int
}

type CheckoutLink struct {
	URL string
}

type InvoiceFile struct {
	FileName      string
	Base64Content string
}

// NewInvoiceFileFromBytes codifica content em base64 pra uso em OrdersClient.ImportInvoice.
func NewInvoiceFileFromBytes(content []byte, fileName string) InvoiceFile {
	return InvoiceFile{FileName: fileName, Base64Content: base64Encode(content)}
}

// OrdersClient expõe as operações de pedidos.
type OrdersClient struct {
	executor                    *transport.Executor
	invoiceUploadAttemptTimeout time.Duration
}

func (c *OrdersClient) Create(ctx context.Context, req CreateOrderRequest) (string, error) {
	wireReq := wire.CreateOrderRequest{
		CPF:                   req.CPF,
		PhoneNumber:           req.PhoneNumber,
		EstablishmentDocument: req.EstablishmentDocument,
		RequestedAmount:       req.RequestedAmount,
		Name:                  req.Name,
		Email:                 req.Email,
		DateOfBirth:           req.DateOfBirth,
		Address: wire.Address{
			Street:       req.Address.Street,
			Number:       req.Address.Number,
			Neighborhood: req.Address.Neighborhood,
			City:         req.Address.City,
			State:        req.Address.State,
			PostalCode:   req.Address.PostalCode,
			Complement:   req.Address.Complement,
		},
	}

	raw, err := c.executor.Post(ctx, "v1/order", wireReq, 0)
	if err != nil {
		return "", translateExecutorError(err)
	}
	if apiErr := transport.EnsureSuccess(raw); apiErr != nil {
		return "", translateExecutorError(apiErr)
	}

	var result wire.IdentifierResponse
	if err := unmarshalJSON(raw.Body, &result); err != nil {
		return "", err
	}
	return result.OrderID, nil
}

func (c *OrdersClient) Get(ctx context.Context, orderID string) (*Order, error) {
	raw, err := c.executor.Get(ctx, fmt.Sprintf("v1/order/%s", orderID))
	if err != nil {
		return nil, translateExecutorError(err)
	}
	if apiErr := transport.EnsureSuccess(raw); apiErr != nil {
		return nil, translateExecutorError(apiErr)
	}

	var result wire.Order
	if err := unmarshalJSON(raw.Body, &result); err != nil {
		return nil, err
	}
	order := orderFromWire(result)
	return &order, nil
}

func (c *OrdersClient) List(ctx context.Context, req ListOrdersRequest) (*PagedResult[Order], error) {
	q := url.Values{}
	if req.Status != nil {
		q.Set("status", strconv.Itoa(int(*req.Status)))
	}
	if req.CustomerDocument != "" {
		q.Set("documentoCliente", req.CustomerDocument)
	}
	if req.StartDate != "" {
		q.Set("dataInicio", req.StartDate)
	}
	if req.EndDate != "" {
		q.Set("dataFim", req.EndDate)
	}
	if req.Number != nil {
		q.Set("numero", strconv.FormatInt(*req.Number, 10))
	}
	if req.EstablishmentDocument != "" {
		q.Set("documentoLoja", req.EstablishmentDocument)
	}
	if req.Description != "" {
		q.Set("descricao", req.Description)
	}
	q.Set("pagina", strconv.Itoa(pageOrDefault(req.Page)))
	q.Set("tamanhoPagina", strconv.Itoa(pageSizeOrDefault(req.PageSize)))

	raw, err := c.executor.Get(ctx, "v1/order/paged?"+q.Encode())
	if err != nil {
		return nil, translateExecutorError(err)
	}
	if apiErr := transport.EnsureSuccess(raw); apiErr != nil {
		return nil, translateExecutorError(apiErr)
	}

	var result wire.PagedOrders
	if err := unmarshalJSON(raw.Body, &result); err != nil {
		return nil, err
	}

	items := make([]Order, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, orderFromWire(item))
	}

	return &PagedResult[Order]{
		Items:       items,
		HasNext:     result.Pagina.HasNext,
		HasPrevious: result.Pagina.HasPrevious,
		PageNumber:  result.Pagina.Number,
		PageSize:    result.Pagina.Size,
		TotalCount:  result.Pagina.Total,
	}, nil
}

func (c *OrdersClient) StartCdcSale(ctx context.Context, orderID string) (*CheckoutLink, error) {
	raw, err := c.executor.Post(ctx, "v1/order/start-cdc-sale", map[string]string{"pedidoId": orderID}, 0)
	if err != nil {
		return nil, translateExecutorError(err)
	}
	if apiErr := transport.EnsureSuccess(raw); apiErr != nil {
		return nil, translateExecutorError(apiErr)
	}

	var result wire.LinkPaymentResponse
	if err := unmarshalJSON(raw.Body, &result); err != nil {
		return nil, err
	}
	return &CheckoutLink{URL: result.Link}, nil
}

func (c *OrdersClient) ImportInvoice(ctx context.Context, orderID string, file InvoiceFile) error {
	wireReq := map[string]string{
		"pedidoId":      orderID,
		"arquivoBase64": file.Base64Content,
		"nomeArquivo":   file.FileName,
	}

	raw, err := c.executor.Post(ctx, "v1/order/invoice", wireReq, c.invoiceUploadAttemptTimeout)
	if err != nil {
		return translateExecutorError(err)
	}
	return translateExecutorError(transport.EnsureSuccess(raw))
}

func orderFromWire(w wire.Order) Order {
	return Order{
		ID:                     w.ID,
		Number:                 w.Number,
		Status:                 orderStatusFromWireValue(w.Status.Value),
		StatusDescription:      w.Status.Description,
		CustomerDocument:       w.CustomerDocument,
		EstablishmentLegalName: w.EstablishmentLegalName,
		EstablishmentDocument:  w.EstablishmentDocument,
		CreatedAt:              w.CreatedAt,
		Total:                  w.Total,
		CustomerName:           w.CustomerName,
		Term:                   w.Term,
		Description:            w.Description,
		ApprovedAmount:         w.ApprovedAmount,
		Disbursed:              w.Disbursed,
		DisbursedAt:            w.DisbursedAt,
		RequestedAmount:        w.RequestedAmount,
	}
}

func orderStatusFromWireValue(value int) OrderStatus {
	switch OrderStatus(value) {
	case OrderStatusUndefined, OrderStatusAnalysing, OrderStatusApproved, OrderStatusUnavailableBalance,
		OrderStatusAnalysisExpired, OrderStatusPendingPayment, OrderStatusBiometryRefused, OrderStatusBiometryApproved,
		OrderStatusPaymentRefused, OrderStatusPurchased, OrderStatusUnauthorized, OrderStatusPendingAuthorization,
		OrderStatusAwaitingRegistration, OrderStatusSaleNotStarted, OrderStatusCanceled, OrderStatusBilling,
		OrderStatusCompleted, OrderStatusFrozen, OrderStatusPendingPaymentConfirmation, OrderStatusDisbursed:
		return OrderStatus(value)
	default:
		return OrderStatusUnknown
	}
}

func pageOrDefault(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

func pageSizeOrDefault(pageSize int) int {
	if pageSize <= 0 {
		return 10
	}
	return pageSize
}
