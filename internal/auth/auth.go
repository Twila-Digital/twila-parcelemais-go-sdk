// Package auth cuida da geração e do cache do token de acesso.
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/httperr"
	"github.com/Twila-Digital/twila-parcelemais-go-sdk/internal/wire"
)

var ErrEmptyTokenResponse = errors.New("a API do Parcele+ retornou uma resposta vazia ao gerar o token de acesso")

// NetworkError marca especificamente uma falha de transporte (não um status HTTP de erro)
// ao chamar o endpoint de token — permite o pacote raiz distinguir isso de uma falha de
// rede genérica numa chamada de recurso (que deve propagar sem tradução, sem virar
// AuthenticationError por engano).
type NetworkError struct {
	Cause error
}

func (e *NetworkError) Error() string { return "falha de rede ao gerar o token de acesso" }
func (e *NetworkError) Unwrap() error { return e.Cause }

type TokenAPIClient struct {
	http    *http.Client
	baseURL string
}

func NewTokenAPIClient(baseURL string, timeout time.Duration) *TokenAPIClient {
	return &TokenAPIClient{http: &http.Client{Timeout: timeout}, baseURL: baseURL}
}

func (c *TokenAPIClient) Generate(ctx context.Context, clientID, clientSecret string) (wire.GenerateAccessTokenResponse, error) {
	reqBody, err := json.Marshal(wire.GenerateAccessTokenRequest{ClientID: clientID, ClientSecret: clientSecret})
	if err != nil {
		return wire.GenerateAccessTokenResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"v1/authentication/accesstoken", bytes.NewReader(reqBody))
	if err != nil {
		return wire.GenerateAccessTokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	raw, err := c.http.Do(req)
	if err != nil {
		return wire.GenerateAccessTokenResponse{}, &NetworkError{Cause: err}
	}
	defer func() { _ = raw.Body.Close() }()

	body, err := io.ReadAll(raw.Body)
	if err != nil {
		return wire.GenerateAccessTokenResponse{}, err
	}

	if raw.StatusCode < 200 || raw.StatusCode >= 300 {
		return wire.GenerateAccessTokenResponse{}, &httperr.Response{StatusCode: raw.StatusCode, Body: body, Header: raw.Header}
	}

	var result wire.GenerateAccessTokenResponse
	if err := json.Unmarshal(body, &result); err != nil || result.Token == "" {
		return wire.GenerateAccessTokenResponse{}, ErrEmptyTokenResponse
	}

	return result, nil
}

const clockSkew = 60 * time.Second

// TokenProvider mantém o token em cache com refresh antecipado, thread-safe via mutex.
type TokenProvider struct {
	client       *TokenAPIClient
	clientID     string
	clientSecret string

	mu        sync.Mutex
	value     string
	expiresAt time.Time
}

func NewTokenProvider(client *TokenAPIClient, clientID, clientSecret string) *TokenProvider {
	return &TokenProvider{client: client, clientID: clientID, clientSecret: clientSecret}
}

func (p *TokenProvider) GetToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.value != "" && !p.isCloseToExpiry() {
		return p.value, nil
	}

	resp, err := p.client.Generate(ctx, p.clientID, p.clientSecret)
	if err != nil {
		return "", err
	}

	p.value = resp.Token
	p.expiresAt = time.Now().Add(time.Duration(resp.ExpiresInS) * time.Second)
	return p.value, nil
}

func (p *TokenProvider) Invalidate() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.value = ""
	p.expiresAt = time.Time{}
}

func (p *TokenProvider) isCloseToExpiry() bool {
	return time.Now().Add(clockSkew).After(p.expiresAt)
}
