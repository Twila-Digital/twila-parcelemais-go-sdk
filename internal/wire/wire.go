// Package wire contém os DTOs que espelham exatamente o formato JSON da API
// do Parcele+ (nomes de campo em português). Nunca são expostos publicamente —
// o pacote raiz sempre mapeia pra/de tipos públicos em inglês.
package wire

type ProblemDetails struct {
	Type          string              `json:"tipo,omitempty"`
	Title         string              `json:"titulo,omitempty"`
	Status        int                 `json:"status,omitempty"`
	Detail        string              `json:"detalhe,omitempty"`
	Instance      string              `json:"instancia,omitempty"`
	Errors        map[string][]string `json:"erros,omitempty"`
	CorrelationID string              `json:"correlationId,omitempty"`
}

type GenerateAccessTokenRequest struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type GenerateAccessTokenResponse struct {
	Token      string `json:"token_de_acesso"`
	ExpiresInS int64  `json:"expira_em_segundos"`
	TokenType  string `json:"tipo_de_token"`
}

type Address struct {
	Street       string `json:"logradouro"`
	Number       string `json:"numero"`
	Neighborhood string `json:"bairro"`
	City         string `json:"cidade"`
	State        string `json:"estado"`
	PostalCode   string `json:"cep"`
	Complement   string `json:"complemento,omitempty"`
}

type CreateOrderRequest struct {
	CPF                   string  `json:"cpf"`
	PhoneNumber           string  `json:"celular"`
	EstablishmentDocument string  `json:"documentoEstabelecimento"`
	RequestedAmount       float64 `json:"valorSolicitado"`
	Name                  string  `json:"nome"`
	Email                 string  `json:"email"`
	DateOfBirth           string  `json:"dataDeNascimento"`
	Address               Address `json:"endereco"`
}

type IdentifierResponse struct {
	OrderID string `json:"pedidoId"`
}

type LinkPaymentResponse struct {
	Link string `json:"linkPagamento"`
}

type OrderStatusWire struct {
	Value       int    `json:"valor"`
	Description string `json:"descricao"`
}

type Order struct {
	ID                     string          `json:"id"`
	Number                 int64           `json:"numero"`
	Status                 OrderStatusWire `json:"status"`
	CustomerDocument       string          `json:"documentoCliente"`
	EstablishmentLegalName string          `json:"razaoSocialEstabelecimento"`
	EstablishmentDocument  string          `json:"documentoEstabelecimento"`
	CreatedAt              string          `json:"criadoEm"`
	Total                  *float64        `json:"total,omitempty"`
	CustomerName           *string         `json:"nomeCliente,omitempty"`
	Term                   *int64          `json:"prazo,omitempty"`
	Description            *string         `json:"descricao,omitempty"`
	ApprovedAmount         *float64        `json:"valorAprovado,omitempty"`
	Disbursed              *bool           `json:"desembolsado,omitempty"`
	DisbursedAt            *string         `json:"desembolsadoEm,omitempty"`
	RequestedAmount        *float64        `json:"valorSolicitado,omitempty"`
}

type Pagina struct {
	HasNext     bool  `json:"tem_proximo"`
	HasPrevious bool  `json:"tem_anterior"`
	Number      int64 `json:"numero"`
	Size        int64 `json:"tamanho"`
	Total       int64 `json:"total"`
}

type PagedOrders struct {
	Items  []Order `json:"itens"`
	Pagina Pagina  `json:"pagina"`
}

type CustomerAddress struct {
	Street       *string `json:"rua,omitempty"`
	Number       *string `json:"numero,omitempty"`
	Neighborhood *string `json:"bairro,omitempty"`
	City         *string `json:"cidade,omitempty"`
	State        *string `json:"estado,omitempty"`
	PostalCode   *string `json:"cep,omitempty"`
	Country      *string `json:"pais,omitempty"`
	Complement   *string `json:"complemento,omitempty"`
}

type Customer struct {
	ID          string           `json:"id"`
	Name        string           `json:"nome"`
	Document    string           `json:"documento"`
	DateOfBirth string           `json:"dataDeNascimento"`
	Address     *CustomerAddress `json:"endereco,omitempty"`
	Email       *string          `json:"email,omitempty"`
	PhoneNumber *string          `json:"celular,omitempty"`
}

type PagedCustomers struct {
	Items  []Customer `json:"itens"`
	Pagina Pagina     `json:"pagina"`
}

type SimulateInstallment struct {
	TotalAmount       float64 `json:"valorTotalDebito"`
	Term              int64   `json:"prazo"`
	InstallmentAmount float64 `json:"valorParcela"`
}

type EstablishmentSimulationValues struct {
	SaleAmount         float64 `json:"valorVenda"`
	DisbursementAmount float64 `json:"valorDesembolso"`
}

type CustomerSimulationValues struct {
	InstallmentAmount float64 `json:"valorParcela"`
}

type SimulationValues struct {
	Establishment EstablishmentSimulationValues `json:"valoresEstabelecimento"`
	Customer      CustomerSimulationValues      `json:"valoresCliente"`
}

type CreateWebhookRequest struct {
	Type               int    `json:"tipo"`
	URL                string `json:"url"`
	AuthenticationType int    `json:"tipoAutenticacao"`
	Credential         string `json:"credencial,omitempty"`
}

type CreateWebhookResponse struct {
	SigningSecret string `json:"chaveAssinatura"`
}

type UpdateWebhookRequest struct {
	URL                string `json:"url"`
	AuthenticationType int    `json:"tipoAutenticacao"`
	Credential         string `json:"credencial,omitempty"`
}

type Webhook struct {
	Type               int    `json:"tipo"`
	URL                string `json:"url"`
	AuthenticationType int    `json:"tipoAutenticacao"`
}

type OrderWebhookEvent struct {
	OrderID    string `json:"id_pedido"`
	StatusEnum int    `json:"enum_status"`
	Status     string `json:"status"`
}
