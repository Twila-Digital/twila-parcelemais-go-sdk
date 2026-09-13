package parcelemais

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/auth"
	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/httperr"
	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/resilience"
	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/wire"
)

func parseProblemDetails(body []byte) ProblemDetails {
	var w wire.ProblemDetails
	if len(body) == 0 {
		return ProblemDetails{}
	}
	if err := json.Unmarshal(body, &w); err != nil {
		return ProblemDetails{}
	}
	return ProblemDetails{
		Type:          w.Type,
		Title:         w.Title,
		Status:        w.Status,
		Detail:        w.Detail,
		Instance:      w.Instance,
		Errors:        w.Errors,
		CorrelationID: w.CorrelationID,
	}
}

func exceptionFromHTTPError(httpErr *httperr.Response) error {
	details := parseProblemDetails(httpErr.Body)
	message := details.Detail
	if message == "" {
		message = details.Title
	}
	if message == "" {
		message = fmt.Sprintf("A API do Parcele+ retornou %d.", httpErr.StatusCode)
	}

	switch {
	case httpErr.StatusCode == 401:
		return &AuthenticationError{Message: message}
	case httpErr.StatusCode == 400 && len(details.Errors) > 0:
		return &ValidationError{APIError: &APIError{StatusCode: 400, Details: details}}
	case httpErr.StatusCode == 429:
		return &RateLimitError{
			APIError:     &APIError{StatusCode: 429, Details: details},
			RetryAfterMs: retryAfterMsFromHeader(httpErr.HeaderValue("Retry-After")),
		}
	default:
		return &APIError{StatusCode: httpErr.StatusCode, Details: details}
	}
}

func retryAfterMsFromHeader(value string) int64 {
	if value == "" {
		return 0
	}
	if seconds, err := strconv.ParseFloat(value, 64); err == nil {
		return int64(seconds * 1000)
	}
	return 0
}

// translateExecutorError converte os erros de baixo nível (internal/resilience,
// internal/httperr, internal/auth) nos tipos públicos deste pacote.
func translateExecutorError(err error) error {
	if err == nil {
		return nil
	}

	var httpErr *httperr.Response
	if errors.As(err, &httpErr) {
		return exceptionFromHTTPError(httpErr)
	}

	var timeoutErr *resilience.TimeoutExceededError
	if errors.As(err, &timeoutErr) {
		return &TimeoutError{Message: timeoutErr.Error(), Cause: err}
	}

	if errors.Is(err, auth.ErrEmptyTokenResponse) {
		return &AuthenticationError{Message: auth.ErrEmptyTokenResponse.Error()}
	}

	var authNetErr *auth.NetworkError
	if errors.As(err, &authNetErr) {
		return &AuthenticationError{Message: "Falha de rede ao gerar o token de acesso.", Cause: authNetErr.Cause}
	}

	// Qualquer outro erro (ex.: falha de rede genuína numa chamada de recurso, depois de
	// esgotar os retries) propaga sem tradução — mesmo comportamento dos SDKs Node/Python/
	// PHP, que deixam o erro de transporte original vazar sem embrulhar num tipo do SDK.
	return err
}
