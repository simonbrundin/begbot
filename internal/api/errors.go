package api

import (
	"encoding/json"
	"net/http"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    string            `json:"code,omitempty"`
	Details []ValidationError `json:"details,omitempty"`
}

func WriteError(w http.ResponseWriter, message string, code string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: message,
		Code:  code,
	})
}

func WriteValidationError(w http.ResponseWriter, details []ValidationError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   "Validation failed",
		Code:    "VALIDATION_ERROR",
		Details: details,
	})
}

func WriteNotFound(w http.ResponseWriter, resource string) {
	WriteError(w, resource+" not found", "NOT_FOUND", 404)
}

func WriteBadRequest(w http.ResponseWriter, message string) {
	WriteError(w, message, "BAD_REQUEST", 400)
}

func WriteServerError(w http.ResponseWriter, message string) {
	WriteError(w, "Internal server error", "INTERNAL_ERROR", 500)
}

func WriteSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(data)
}
