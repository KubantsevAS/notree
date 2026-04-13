package httputil

import (
	"encoding/json"
	"net/http"
)

func WriteResponseJSON(w http.ResponseWriter, data any, StatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(StatusCode)
	json.NewEncoder(w).Encode(data)
}
