package httptransport

import (
	"encoding/json"
	"net/http"
)

type AppError struct {
	Status  int
	Code    string
	Message string
	Details any
}

func (e *AppError) Error() string {
	return e.Code
}

func writeError(w http.ResponseWriter, err *AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code":    err.Code,
		"message": err.Message,
		"details": err.Details,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
