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

	writeBody(w, errorResponse{status, message})

	w.WriteHeader(status)
}

type successResponse struct {
	Status  int
	Message string
}

func Ok(w http.ResponseWriter) {
	writeBody(w, successResponse{http.StatusOK, "OK"})

	w.WriteHeader(http.StatusOK)
}

func Success(w http.ResponseWriter, body any) {

	if body == nil {
		Ok(w)
		return
	}

	writeBody(w, body)

	w.WriteHeader(http.StatusOK)
}

func writeBody(w http.ResponseWriter, body any) {
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("failed to write response body", "error", err)
	}
}
