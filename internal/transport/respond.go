package transport

import (
	"encoding/json"
	"net/http"
)

// errorResponse — единый формат ошибки API.
type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// respondError отправляет клиенту JSON с ошибкой.
func RespondError(w http.ResponseWriter, status int, code, message string) {
	RespondJSON(w, status, errorResponse{
		Error: errorBody{
			Code:    code,
			Message: message,
		},
	})
}

// respondJSON отправляет клиенту JSON с произвольным телом.
func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}
