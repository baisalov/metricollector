package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type errorResponse struct {
	Status int
	Error  string
}

func Error(w http.ResponseWriter, message string, status int) {
	w.WriteHeader(status)

	writeBody(w, errorResponse{status, message})
}

type successResponse struct {
	Status  int
	Message string
}

func Ok(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)

	writeBody(w, successResponse{http.StatusOK, "OK"})
}

func Success(w http.ResponseWriter, body any) {

	if body == nil {
		Ok(w)
		return
	}

	w.WriteHeader(http.StatusOK)

	writeBody(w, body)
}

func writeBody(w http.ResponseWriter, body any) {
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("failed to write response body", "error", err)
	}
}
