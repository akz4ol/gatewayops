// Package handler provides HTTP handlers for the gateway.
package handler

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SuccessResponse represents a successful response.
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// WriteJSON writes a JSON response.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteError writes an error response.
func WriteError(w http.ResponseWriter, status int, code string, message string) {
	WriteJSON(w, status, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// WriteSuccess writes a success response.
func WriteSuccess(w http.ResponseWriter, status int, data interface{}) {
	WriteJSON(w, status, SuccessResponse{
		Data: data,
	})
}
