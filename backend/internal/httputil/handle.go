package httputil

import (
	"net/http"
)

func HandleBody[T any](r *http.Request) (*T, error) {
	body, err := Decode[T](r.Body)
	if err != nil {
		return nil, err
	}

	if err = IsValid(body); err != nil {
		return nil, err
	}

	return &body, nil
}
