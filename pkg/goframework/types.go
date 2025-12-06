package goframework

import (
	"fmt"
)

// HandlerFunc é a assinatura padrão para handlers das rotas.
type HandlerFunc func(*Context) error

// HTTPError representa um erro com código HTTP específico.
type HTTPError struct {
	Status  int
	Message string
}

func (e *HTTPError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%d %s", e.Status, e.Message)
}

// NewHTTPError cria um erro HTTP.
func NewHTTPError(status int, message string) error {
	return &HTTPError{
		Status:  status,
		Message: message,
	}
}
