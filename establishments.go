package parcelemais

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/transport"
	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/wire"
)

type DisbursementModel int

const (
	DisbursementModelEstablishmentChain DisbursementModel = 1
	DisbursementModelEstablishment      DisbursementModel = 2
	DisbursementModelExternal           DisbursementModel = 3
)

type BankAccountType int

const (
	BankAccountTypeCurrent BankAccountType = 1
	BankAccountTypeSavings BankAccountType = 2
	BankAccountTypePayment BankAccountType = 3
)

type EstablishmentOwner struct {
	Name  string
	Email string
	Phone string
}

type EstablishmentBankAccount struct {
	BankNumber     string
	AgencyNumber   string
	AgencyDigit    string
	AccountNumber  string
	AccountDigit   string
	AccountType    BankAccountType
	HolderName     string
	HolderDocument string
}

type EstablishmentAddress struct {
	Street     string
	Number     string
	Complement string
	District   string
	City       string
	State      string
	ZipCode    string
	Country    string
}

type CreateEstablishmentRequest struct {
	Document          string
	LegalName         string
	TradeName         string
	DisbursementModel DisbursementModel
	Owner             EstablishmentOwner
	BankAccount       EstablishmentBankAccount
	Address           *EstablishmentAddress
}

type CreateEstablishmentResult struct {
	EstablishmentID string
}

type UpdateEstablishmentRequest struct {
	TradeName         string
	DisbursementModel *DisbursementModel
	Address           *EstablishmentAddress
}

type Establishment struct {
	EstablishmentID   string
	Document          string
	LegalName         string
	TradeName         string
	IsActive          bool
	Owner             EstablishmentOwner
	DisbursementModel *DisbursementModel
	BankAccount       *EstablishmentBankAccount
	Address           *EstablishmentAddress
}

type ListEstablishmentsRequest struct {
	TradeName string
	IsActive  *bool
}

type EstablishmentsClient struct {
	executor *transport.Executor
}

func (c *EstablishmentsClient) Create(ctx context.Context, req CreateEstablishmentRequest) (*CreateEstablishmentResult, error) {
	wireReq := wire.CreateEstablishmentRequest{
		Document:          req.Document,
		LegalName:         req.LegalName,
		TradeName:         req.TradeName,
		DisbursementModel: int(req.DisbursementModel),
		Owner:             wire.EstablishmentOwner{Name: req.Owner.Name, Email: req.Owner.Email, Phone: req.Owner.Phone},
		BankAccount:       bankAccountToWire(req.BankAccount),
		Address:           addressToWire(req.Address),
	}

	raw, err := c.executor.Post(ctx, "v1/establishment", wireReq, 0)
	if err != nil {
		return nil, translateExecutorError(err)
	}
	if apiErr := transport.EnsureSuccess(raw); apiErr != nil {
		return nil, translateExecutorError(apiErr)
	}

	var result wire.CreateEstablishmentResponse
	if err := unmarshalJSON(raw.Body, &result); err != nil {
		return nil, err
	}
	return &CreateEstablishmentResult{EstablishmentID: result.EstablishmentID}, nil
}

func (c *EstablishmentsClient) Get(ctx context.Context, establishmentID string) (*Establishment, error) {
	raw, err := c.executor.Get(ctx, fmt.Sprintf("v1/establishment/%s", establishmentID))
	if err != nil {
		return nil, translateExecutorError(err)
	}
	if apiErr := transport.EnsureSuccess(raw); apiErr != nil {
		return nil, translateExecutorError(apiErr)
	}

	var result wire.Establishment
	if err := unmarshalJSON(raw.Body, &result); err != nil {
		return nil, err
	}
	establishment := establishmentFromWire(result)
	return &establishment, nil
}

func (c *EstablishmentsClient) List(ctx context.Context, req ListEstablishmentsRequest) ([]Establishment, error) {
	path := "v1/establishment/list"

	q := url.Values{}
	if req.TradeName != "" {
		q.Set("nomeFantasia", req.TradeName)
	}
	if req.IsActive != nil {
		q.Set("ativa", strconv.FormatBool(*req.IsActive))
	}
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	raw, err := c.executor.Get(ctx, path)
	if err != nil {
		return nil, translateExecutorError(err)
	}
	if apiErr := transport.EnsureSuccess(raw); apiErr != nil {
		return nil, translateExecutorError(apiErr)
	}

	var result []wire.Establishment
	if err := unmarshalJSON(raw.Body, &result); err != nil {
		return nil, err
	}

	establishments := make([]Establishment, 0, len(result))
	for _, item := range result {
		establishments = append(establishments, establishmentFromWire(item))
	}
	return establishments, nil
}

func (c *EstablishmentsClient) Update(ctx context.Context, establishmentID string, req UpdateEstablishmentRequest) error {
	wireReq := wire.UpdateEstablishmentRequest{
		TradeName: req.TradeName,
		Address:   addressToWire(req.Address),
	}

	if req.DisbursementModel != nil {
		disbursementModel := int(*req.DisbursementModel)
		wireReq.DisbursementModel = &disbursementModel
	}

	raw, err := c.executor.Put(ctx, fmt.Sprintf("v1/establishment/%s", establishmentID), wireReq)
	if err != nil {
		return translateExecutorError(err)
	}
	return translateExecutorError(transport.EnsureSuccess(raw))
}

func (c *EstablishmentsClient) UpdateBankAccount(ctx context.Context, establishmentID string, bankAccount EstablishmentBankAccount) error {
	raw, err := c.executor.Put(ctx, fmt.Sprintf("v1/establishment/%s/bank-account", establishmentID), bankAccountToWire(bankAccount))
	if err != nil {
		return translateExecutorError(err)
	}
	return translateExecutorError(transport.EnsureSuccess(raw))
}

func (c *EstablishmentsClient) Activate(ctx context.Context, establishmentID string) error {
	return c.setActive(ctx, establishmentID, true)
}

func (c *EstablishmentsClient) Deactivate(ctx context.Context, establishmentID string) error {
	return c.setActive(ctx, establishmentID, false)
}

func (c *EstablishmentsClient) setActive(ctx context.Context, establishmentID string, isActive bool) error {
	raw, err := c.executor.Put(ctx, fmt.Sprintf("v1/establishment/%s/status", establishmentID), wire.UpdateEstablishmentStatusRequest{IsActive: isActive})
	if err != nil {
		return translateExecutorError(err)
	}
	return translateExecutorError(transport.EnsureSuccess(raw))
}

func bankAccountToWire(bankAccount EstablishmentBankAccount) wire.EstablishmentBankAccount {
	return wire.EstablishmentBankAccount{
		BankNumber:     bankAccount.BankNumber,
		AgencyNumber:   bankAccount.AgencyNumber,
		AgencyDigit:    bankAccount.AgencyDigit,
		AccountNumber:  bankAccount.AccountNumber,
		AccountDigit:   bankAccount.AccountDigit,
		AccountType:    int(bankAccount.AccountType),
		HolderName:     bankAccount.HolderName,
		HolderDocument: bankAccount.HolderDocument,
	}
}

func addressToWire(address *EstablishmentAddress) *wire.EstablishmentAddress {
	if address == nil {
		return nil
	}

	return &wire.EstablishmentAddress{
		Street:     address.Street,
		Number:     address.Number,
		Complement: address.Complement,
		District:   address.District,
		City:       address.City,
		State:      address.State,
		ZipCode:    address.ZipCode,
		Country:    address.Country,
	}
}

func establishmentFromWire(w wire.Establishment) Establishment {
	establishment := Establishment{
		EstablishmentID: w.EstablishmentID,
		Document:        w.Document,
		LegalName:       w.LegalName,
		TradeName:       w.TradeName,
		IsActive:        w.IsActive,
		Owner:           EstablishmentOwner{Name: w.Owner.Name, Email: w.Owner.Email, Phone: w.Owner.Phone},
	}

	if w.DisbursementModel != nil {
		disbursementModel := DisbursementModel(*w.DisbursementModel)
		establishment.DisbursementModel = &disbursementModel
	}

	if w.BankAccount != nil {
		establishment.BankAccount = &EstablishmentBankAccount{
			BankNumber:     w.BankAccount.BankNumber,
			AgencyNumber:   w.BankAccount.AgencyNumber,
			AgencyDigit:    w.BankAccount.AgencyDigit,
			AccountNumber:  w.BankAccount.AccountNumber,
			AccountDigit:   w.BankAccount.AccountDigit,
			AccountType:    BankAccountType(w.BankAccount.AccountType),
			HolderName:     w.BankAccount.HolderName,
			HolderDocument: w.BankAccount.HolderDocument,
		}
	}

	if w.Address != nil {
		establishment.Address = &EstablishmentAddress{
			Street:     w.Address.Street,
			Number:     w.Address.Number,
			Complement: w.Address.Complement,
			District:   w.Address.District,
			City:       w.Address.City,
			State:      w.Address.State,
			ZipCode:    w.Address.ZipCode,
			Country:    w.Address.Country,
		}
	}

	return establishment
}
