package parcelemais

import (
	"context"
	"net/url"
	"strconv"

	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/transport"
	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/wire"
)

type CalculationValueType int

const (
	CalculationValueTypeGrossAmount  CalculationValueType = 1
	CalculationValueTypeLiquidAmount CalculationValueType = 2
)

type SimulateInstallmentsRequest struct {
	RequestedAmount      float64
	CalculationValueType CalculationValueType // zero-value trata como GrossAmount
}

type SimulateValuesRequest struct {
	Amount               float64
	Term                 int64
	CalculationValueType CalculationValueType
}

type InstallmentSimulation struct {
	TotalAmount       float64
	Term              int64
	InstallmentAmount float64
}

type ValuesSimulation struct {
	SaleAmount         float64
	DisbursementAmount float64
	InstallmentAmount  float64
}

type SimulationsClient struct {
	executor *transport.Executor
}

func (c *SimulationsClient) SimulateInstallments(ctx context.Context, req SimulateInstallmentsRequest) ([]InstallmentSimulation, error) {
	calcType := req.CalculationValueType
	if calcType == 0 {
		calcType = CalculationValueTypeGrossAmount
	}

	q := url.Values{}
	q.Set("valorSolicitado", strconv.FormatFloat(req.RequestedAmount, 'f', -1, 64))
	q.Set("tipoValorCalculo", strconv.Itoa(int(calcType)))

	raw, err := c.executor.Get(ctx, "v1/order/simulate-installments?"+q.Encode())
	if err != nil {
		return nil, translateExecutorError(err)
	}
	if apiErr := transport.EnsureSuccess(raw); apiErr != nil {
		return nil, translateExecutorError(apiErr)
	}

	var result []wire.SimulateInstallment
	if err := unmarshalJSON(raw.Body, &result); err != nil {
		return nil, err
	}

	installments := make([]InstallmentSimulation, 0, len(result))
	for _, item := range result {
		installments = append(installments, InstallmentSimulation{
			TotalAmount:       item.TotalAmount,
			Term:              item.Term,
			InstallmentAmount: item.InstallmentAmount,
		})
	}
	return installments, nil
}

func (c *SimulationsClient) SimulateValues(ctx context.Context, req SimulateValuesRequest) (*ValuesSimulation, error) {
	calcType := req.CalculationValueType
	if calcType == 0 {
		calcType = CalculationValueTypeGrossAmount
	}

	q := url.Values{}
	q.Set("valor", strconv.FormatFloat(req.Amount, 'f', -1, 64))
	q.Set("prazo", strconv.FormatInt(req.Term, 10))
	q.Set("modeloJuros", "1")
	q.Set("tipoValorCalculo", strconv.Itoa(int(calcType)))

	raw, err := c.executor.Get(ctx, "v1/order/simulate-values?"+q.Encode())
	if err != nil {
		return nil, translateExecutorError(err)
	}
	if apiErr := transport.EnsureSuccess(raw); apiErr != nil {
		return nil, translateExecutorError(apiErr)
	}

	var result wire.SimulationValues
	if err := unmarshalJSON(raw.Body, &result); err != nil {
		return nil, err
	}

	return &ValuesSimulation{
		SaleAmount:         result.Establishment.SaleAmount,
		DisbursementAmount: result.Establishment.DisbursementAmount,
		InstallmentAmount:  result.Customer.InstallmentAmount,
	}, nil
}
