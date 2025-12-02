package http

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrResponse represents an error response
type ErrResponse struct {
	Code  int64  `json:"code,omitempty"`
	Error string `json:"error,omitempty"`
}

// WriteError writes an error response with the given status code
func WriteError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if encErr := json.NewEncoder(w).Encode(ErrResponse{Error: err.Error()}); encErr != nil {
		log.Printf("failed to encode error response: %v", encErr)
	}
}

// ErrInternal writes an internal server error response
func ErrInternal(w http.ResponseWriter, err error) {
	WriteError(w, http.StatusInternalServerError, err)
}

// ErrBadRequest writes a bad request error response
func ErrBadRequest(w http.ResponseWriter, err error) {
	WriteError(w, http.StatusBadRequest, err)
}

// ErrNotFound writes a not found error response
func ErrNotFound(w http.ResponseWriter, err error) {
	WriteError(w, http.StatusNotFound, err)
}
