// Package transport executa requisições HTTP autenticadas com idempotency key
// automática e pipeline de resiliência (retry + circuit breaker + timeout total).
package transport

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/httperr"
	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/idempotency"
	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/resilience"
)

// TokenProvider é satisfeito estruturalmente por auth.TokenProvider — sem import direto
// entre os dois pacotes.
type TokenProvider interface {
	GetToken(ctx context.Context) (string, error)
	Invalidate()
}

type RawResponse struct {
	StatusCode int
	Body       []byte
	Header     http.Header
}

type Executor struct {
	http                   *http.Client
	baseURL                string
	tokenProvider          TokenProvider
	resiliencePipeline     *resilience.Pipeline
	attemptTimeout         time.Duration
	disableAutoIdempotency bool
}

func NewExecutor(
	baseURL string,
	tokenProvider TokenProvider,
	resilienceOpts resilience.Options,
	attemptTimeout time.Duration,
	disableAutoIdempotency bool,
) *Executor {
	return &Executor{
		http:                   &http.Client{},
		baseURL:                baseURL,
		tokenProvider:          tokenProvider,
		resiliencePipeline:     resilience.NewPipeline(resilienceOpts),
		attemptTimeout:         attemptTimeout,
		disableAutoIdempotency: disableAutoIdempotency,
	}
}

func (e *Executor) Get(ctx context.Context, path string) (RawResponse, error) {
	return e.send(ctx, http.MethodGet, path, nil, e.attemptTimeout)
}

func (e *Executor) Post(ctx context.Context, path string, body interface{}, attemptTimeout time.Duration) (RawResponse, error) {
	if attemptTimeout <= 0 {
		attemptTimeout = e.attemptTimeout
	}
	return e.send(ctx, http.MethodPost, path, body, attemptTimeout)
}

func (e *Executor) Put(ctx context.Context, path string, body interface{}) (RawResponse, error) {
	return e.send(ctx, http.MethodPut, path, body, e.attemptTimeout)
}

func (e *Executor) Delete(ctx context.Context, path string) (RawResponse, error) {
	return e.send(ctx, http.MethodDelete, path, nil, e.attemptTimeout)
}

// EnsureSuccess retorna um *httperr.Response (como error) se o status não for 2xx.
func EnsureSuccess(resp RawResponse) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &httperr.Response{StatusCode: resp.StatusCode, Body: resp.Body, Header: resp.Header}
	}
	return nil
}

func (e *Executor) send(ctx context.Context, method, path string, body interface{}, attemptTimeout time.Duration) (RawResponse, error) {
	var idempotencyKey string
	if !e.disableAutoIdempotency && idempotency.RequiresKey(method, path) {
		idempotencyKey = newUUIDv4()
	}
	retrySafe := idempotency.IsRetrySafe(method, path, idempotencyKey != "")

	var bodyBytes []byte
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return RawResponse{}, err
		}
		bodyBytes = encoded
	}

	var lastRaw RawResponse

	_, err := e.resiliencePipeline.Execute(ctx, retrySafe, isNetworkError, func(attemptCtx context.Context) (resilience.Response, error) {
		raw, sendErr := e.sendWithAuth(attemptCtx, method, path, bodyBytes, idempotencyKey, attemptTimeout)
		if sendErr != nil {
			return resilience.Response{}, sendErr
		}
		lastRaw = raw
		return resilience.Response{StatusCode: raw.StatusCode, RetryAfter: parseRetryAfter(raw.Header)}, nil
	})

	if err != nil {
		var timeoutErr *resilience.TimeoutExceededError
		if errors.As(err, &timeoutErr) {
			return RawResponse{}, timeoutErr
		}
		return RawResponse{}, err
	}

	return lastRaw, nil
}

func (e *Executor) sendWithAuth(ctx context.Context, method, path string, body []byte, idempotencyKey string, attemptTimeout time.Duration) (RawResponse, error) {
	token, err := e.tokenProvider.GetToken(ctx)
	if err != nil {
		return RawResponse{}, err
	}

	resp, err := e.sendOnce(ctx, method, path, body, idempotencyKey, attemptTimeout, token)
	if err != nil {
		return RawResponse{}, err
	}
	if resp.StatusCode != http.StatusUnauthorized {
		return resp, nil
	}

	e.tokenProvider.Invalidate()
	newToken, err := e.tokenProvider.GetToken(ctx)
	if err != nil {
		return RawResponse{}, err
	}

	retried, err := e.sendOnce(ctx, method, path, body, idempotencyKey, attemptTimeout, newToken)
	if err != nil {
		return RawResponse{}, err
	}
	return retried, nil
}

func (e *Executor) sendOnce(ctx context.Context, method, path string, body []byte, idempotencyKey string, attemptTimeout time.Duration, token string) (RawResponse, error) {
	// baseURL sempre termina com "/" (garantido na resolução de ClientOptions) e path nunca
	// começa com "/" — concatenação direta evita a armadilha de resolução de URI relativa
	// que descartaria o segmento de ambiente da base URL.
	fullURL := e.baseURL + path

	attemptCtx := ctx
	var cancel context.CancelFunc
	if attemptTimeout > 0 {
		attemptCtx, cancel = context.WithTimeout(ctx, attemptTimeout)
		defer cancel()
	}

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(attemptCtx, method, fullURL, reader)
	if err != nil {
		return RawResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	raw, err := e.http.Do(req)
	if err != nil {
		return RawResponse{}, err
	}
	defer func() { _ = raw.Body.Close() }()

	respBody, err := io.ReadAll(raw.Body)
	if err != nil {
		return RawResponse{}, err
	}

	return RawResponse{StatusCode: raw.StatusCode, Body: respBody, Header: raw.Header}, nil
}

func isNetworkError(err error) bool {
	var httpErr *httperr.Response
	if errors.As(err, &httpErr) {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var urlErr *url.Error
	return errors.As(err, &urlErr)
}

func parseRetryAfter(header http.Header) time.Duration {
	value := header.Get("Retry-After")
	if value == "" {
		return 0
	}
	if seconds, err := strconv.ParseFloat(value, 64); err == nil {
		return time.Duration(seconds * float64(time.Second))
	}
	if when, err := http.ParseTime(value); err == nil {
		if d := time.Until(when); d > 0 {
			return d
		}
	}
	return 0
}

func newUUIDv4() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16])
}
