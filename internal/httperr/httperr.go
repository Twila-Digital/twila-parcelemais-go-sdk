// Package httperr representa uma resposta HTTP não-2xx como error, sem depender de
// tipos do pacote raiz — usado pelos pacotes internos (auth, transport) pra propagar
// o status/corpo da resposta pra cima, onde o pacote raiz reconstrói (via errors.As)
// o erro tipado público correto (AuthenticationError, ValidationError, etc.).
package httperr

import (
	"fmt"
	"net/http"
	"strings"
)

type Response struct {
	StatusCode int
	Body       []byte
	Header     http.Header
}

func (r *Response) Error() string {
	return fmt.Sprintf("parcelemais: http status %d", r.StatusCode)
}

func (r *Response) HeaderValue(name string) string {
	for k, v := range r.Header {
		if strings.EqualFold(k, name) && len(v) > 0 {
			return v[0]
		}
	}
	return ""
}
