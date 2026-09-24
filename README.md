<p align="center">
  <img src="https://raw.githubusercontent.com/Twila-Digital/twila-parcelemais-go-sdk/production/assets/logo-light.svg" alt="Parcele+" width="180" style="max-width: 100%;">
</p>

<p align="center">
  <a href="LICENSE"><img alt="License" src="https://img.shields.io/github/license/Twila-Digital/twila-parcelemais-go-sdk"></a>
  <a href="https://github.com/Twila-Digital/twila-parcelemais-go-sdk/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/Twila-Digital/twila-parcelemais-go-sdk/actions/workflows/ci.yml/badge.svg"></a>
  <a href="https://github.com/Twila-Digital/twila-parcelemais-go-sdk/actions/workflows/quality.yml"><img alt="Quality" src="https://github.com/Twila-Digital/twila-parcelemais-go-sdk/actions/workflows/quality.yml/badge.svg"></a>
  <a href="https://github.com/Twila-Digital/twila-parcelemais-go-sdk/security/code-scanning"><img alt="Security" src="https://github.com/Twila-Digital/twila-parcelemais-go-sdk/actions/workflows/security.yml/badge.svg"></a>
  <a href="https://codecov.io/gh/Twila-Digital/twila-parcelemais-go-sdk"><img alt="Coverage" src="https://codecov.io/gh/Twila-Digital/twila-parcelemais-go-sdk/branch/production/graph/badge.svg"></a>
  <a href="https://pkg.go.dev/github.com/Twila-Digital/twila-parcelemais-go-sdk"><img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/Twila-Digital/twila-parcelemais-go-sdk.svg"></a>
</p>

# twila-parcelemais-go-sdk

SDK oficial em Go para a API do [Parcele+](https://www.cartaosimples.com.br) — crédito direto ao consumidor (CDC) e parcelamento no momento da compra.

> Uso restrito a server-side. O `ClientSecret` nunca deve ser embarcado em um app mobile, frontend ou qualquer código que rode no navegador/dispositivo do usuário final.

## Compatibilidade

| Runtime | Versões aceitas |
| --- | --- |
| Go | 1.18 ou superior (CI cobre 1.18, 1.20, 1.21, 1.22, 1.23, 1.24, 1.25, 1.26 e 1.27) |

**Zero dependências externas** — só biblioteca padrão (`net/http`, `encoding/json`, `crypto/hmac`, `testing`). Seguro para uso concorrente por múltiplas goroutines.

## Instalação

```bash
go get github.com/Twila-Digital/twila-parcelemais-go-sdk
```

## Configuração

```go
import parcelemais "github.com/Twila-Digital/twila-parcelemais-go-sdk"

client, err := parcelemais.NewClient(parcelemais.ClientOptions{
    ClientID:     "<seu-client-id>",
    ClientSecret: "<seu-client-secret>",
    Environment:  parcelemais.EnvironmentStaging,
})
```

`*parcelemais.Client` deve ser reaproveitado como singleton na sua aplicação (não crie uma instância por requisição) — ele mantém o cache do token de acesso e o estado do circuit breaker, e é seguro para uso concorrente.

## Uso

```go
parcelas, err := client.Simulations.SimulateInstallments(ctx, parcelemais.SimulateInstallmentsRequest{
    RequestedAmount: 1500.0,
})
for _, parcela := range parcelas {
    fmt.Printf("%dx de %.2f (total %.2f)\n", parcela.Term, parcela.InstallmentAmount, parcela.TotalAmount)
}
```

O client expõe um struct por recurso:

| Cliente | Métodos |
| --- | --- |
| `client.Orders` | `Create`, `Get`, `List`, `StartCdcSale`, `ImportInvoice` |
| `client.Simulations` | `SimulateInstallments`, `SimulateValues` |
| `client.Customers` | `Get`, `List` |
| `client.Establishments` | `Create`, `Get`, `List`, `Update`, `UpdateBankAccount`, `Activate`, `Deactivate` |
| `client.Webhooks` | `Create`, `List`, `ListAudit`, `Update`, `Delete` |

Todo método recebe `context.Context` como primeiro argumento (idiomático em Go, permite cancelamento/timeout por chamada além do timeout total configurado no client).

## Paginação

`Orders.List(...)`, `Customers.List(...)` e `Webhooks.ListAudit(...)` retornam `*PagedResult[T]` (genérico) — sem auto-paginação, você controla explicitamente o avanço de página:

```go
page, err := client.Orders.List(ctx, parcelemais.ListOrdersRequest{Page: 1, PageSize: 20})
for _, order := range page.Items {
    fmt.Println(order.ID)
}
if page.HasNext {
    next, err := client.Orders.List(ctx, parcelemais.ListOrdersRequest{Page: 2, PageSize: 20})
}
```

## Tratamento de erros

Segue o idioma Go — erros são valores retornados, não exceções. Use `errors.As` para verificar o tipo:

| Erro | Quando |
| --- | --- |
| `*parcelemais.ConfigurationError` | Configuração do `ClientOptions` inválida (ex.: `ClientID`/`ClientSecret` ausentes) |
| `*parcelemais.AuthenticationError` | Falha ao gerar/renovar o token de acesso |
| `*parcelemais.ValidationError` | `400` — erro de validação, com `FieldErrors()` por campo |
| `*parcelemais.RateLimitError` | `429` |
| `*parcelemais.TimeoutError` | Timeout de rede, timeout total, ou circuit breaker aberto |
| `*parcelemais.APIError` | Qualquer outro erro de API (`404`, `409`, `5xx`) |
| `*parcelemais.WebhookSignatureError` | Assinatura de webhook inválida ou expirada |

```go
import "errors"

order, err := client.Orders.Get(ctx, orderID)
var apiErr *parcelemais.APIError
if errors.As(err, &apiErr) {
    fmt.Printf("%d %s: %s\n", apiErr.StatusCode, apiErr.ErrorCode(), apiErr)
}
```

## Validando webhooks

```go
event, err := parcelemais.ParseWebhookEvent(rawBody, signatureHeader, signingSecret)
```

Verifica a assinatura HMAC-SHA256 do cabeçalho e a janela de replay (5 minutos) antes de expor o evento. Retorna `*parcelemais.WebhookSignatureError` se a assinatura for inválida ou o evento estiver fora da janela.

## Samples

- `samples/sample-cli` — script standalone, sem framework
- `samples/sample-http` — servidor HTTP com `net/http` puro (sem framework — a stdlib do Go já cobre isso)

## Qualidade, segurança e cobertura

- **Build/Test** (`ci.yml`) — `go vet` + `golangci-lint` + `gofmt -l` no job build; testes (`go test -race`) em Go 1.18–1.27.
- **Quality** (`quality.yml`) — análise estática via Codacy CLI (gometalinter/staticcheck), resultados publicados na aba **Security → Code scanning** do repositório.
- **Security** (`security.yml`) — [CodeQL](https://codeql.github.com/) para Go, rodando a cada PR/push e semanalmente.
- **Coverage** — cobertura de testes coletada via `go test -coverprofile` e publicada no [Codecov](https://codecov.io/gh/Twila-Digital/twila-parcelemais-go-sdk).

## Documentação completa

[documentacao.parcelemais.com.br](https://documentacao.parcelemais.com.br) — referência de todos os endpoints, autenticação, webhooks e mais.

## Contribuindo

Veja [CONTRIBUTING.md](CONTRIBUTING.md).

## Código de conduta

Este projeto segue o [Código de Conduta](CODE_OF_CONDUCT.md).

## Licença

[MIT](LICENSE)
