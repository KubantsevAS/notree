package httputil

import (
	"net/http"

	"github.com/KubantsevAS/notree/backend/internal/validator"
)

func HandleBody[T any](r *http.Request) (*T, error) {
	body, err := Decode[T](r.Body)
	if err != nil {
		return nil, err
	}

	if err = validator.Check(body); err != nil {
		return nil, err
	}

	return &body, nil
}
