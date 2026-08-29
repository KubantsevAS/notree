package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/KubantsevAS/notree/backend/internal/http/dto"
)

func WriteResponseJSON(w http.ResponseWriter, data any, StatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(StatusCode)
	json.NewEncoder(w).Encode(data)
}

func WriteErrorJSON(w http.ResponseWriter, message string, StatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(StatusCode)
	json.NewEncoder(w).Encode(dto.ErrorResponse{Error: message})
}
