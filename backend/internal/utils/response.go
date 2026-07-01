package utils

import (
	"encoding/json"
	"net/http"

	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/dto"
)

const (
	StatusSuccess = "success"
	StatusError   = "error"
)

// SendSuccess writes a successful JSON response using the standard envelope.
func SendSuccess(w http.ResponseWriter, statusCode int, message string, data any) error {
	return sendJSON(w, statusCode, dto.Response{
		Status:  StatusSuccess,
		Message: message,
		Data:    data,
		Errors:  nil,
	})
}

// SendError writes an error JSON response using the standard envelope.
func SendError(w http.ResponseWriter, statusCode int, message string, errors map[string]string) error {
	return sendJSON(w, statusCode, dto.Response{
		Status:  StatusError,
		Message: message,
		Data:    nil,
		Errors:  errors,
	})
}

// DecodeJSON decodes a JSON request body into target.
func DecodeJSON(r *http.Request, target any) error {
	return json.NewDecoder(r.Body).Decode(target)
}

func sendJSON(w http.ResponseWriter, statusCode int, response dto.Response) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(response)
}
