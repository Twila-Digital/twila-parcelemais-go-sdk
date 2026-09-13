package parcelemais

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/transport"
	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/wire"
)

type CustomerAddress struct {
	Street       *string
	Number       *string
	Neighborhood *string
	City         *string
	State        *string
	PostalCode   *string
	Country      *string
	Complement   *string
}

type Customer struct {
	ID          string
	Name        string
	Document    string
	DateOfBirth string
	Address     *CustomerAddress
	Email       *string
	PhoneNumber *string
}

type ListCustomersRequest struct {
	Name     string
	Document string
	Page     int
	PageSize int
}

type CustomersClient struct {
	executor *transport.Executor
}

func (c *CustomersClient) Get(ctx context.Context, customerID string) (*Customer, error) {
	raw, err := c.executor.Get(ctx, fmt.Sprintf("v1/customer/%s", customerID))
	if err != nil {
		return nil, translateExecutorError(err)
	}
	if apiErr := transport.EnsureSuccess(raw); apiErr != nil {
		return nil, translateExecutorError(apiErr)
	}

	var result wire.Customer
	if err := unmarshalJSON(raw.Body, &result); err != nil {
		return nil, err
	}
	customer := customerFromWire(result)
	return &customer, nil
}

func (c *CustomersClient) List(ctx context.Context, req ListCustomersRequest) (*PagedResult[Customer], error) {
	q := url.Values{}
	if req.Name != "" {
		q.Set("nome", req.Name)
	}
	if req.Document != "" {
		q.Set("documento", req.Document)
	}
	q.Set("pagina", strconv.Itoa(pageOrDefault(req.Page)))
	q.Set("tamanhoPagina", strconv.Itoa(pageSizeOrDefault(req.PageSize)))

	raw, err := c.executor.Get(ctx, "v1/customer/paged?"+q.Encode())
	if err != nil {
		return nil, translateExecutorError(err)
	}
	if apiErr := transport.EnsureSuccess(raw); apiErr != nil {
		return nil, translateExecutorError(apiErr)
	}

	var result wire.PagedCustomers
	if err := unmarshalJSON(raw.Body, &result); err != nil {
		return nil, err
	}

	items := make([]Customer, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, customerFromWire(item))
	}

	return &PagedResult[Customer]{
		Items:       items,
		HasNext:     result.Pagina.HasNext,
		HasPrevious: result.Pagina.HasPrevious,
		PageNumber:  result.Pagina.Number,
		PageSize:    result.Pagina.Size,
		TotalCount:  result.Pagina.Total,
	}, nil
}

func customerFromWire(w wire.Customer) Customer {
	var address *CustomerAddress
	if w.Address != nil {
		address = &CustomerAddress{
			Street:       w.Address.Street,
			Number:       w.Address.Number,
			Neighborhood: w.Address.Neighborhood,
			City:         w.Address.City,
			State:        w.Address.State,
			PostalCode:   w.Address.PostalCode,
			Country:      w.Address.Country,
			Complement:   w.Address.Complement,
		}
	}

	return Customer{
		ID:          w.ID,
		Name:        w.Name,
		Document:    w.Document,
		DateOfBirth: w.DateOfBirth,
		Address:     address,
		Email:       w.Email,
		PhoneNumber: w.PhoneNumber,
	}
}
