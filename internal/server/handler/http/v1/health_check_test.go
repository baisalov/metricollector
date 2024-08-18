package v1

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Mock implementation of checker interface for testing purposes
type mockChecker struct {
	checkFunc func(ctx context.Context) error
}

func (m *mockChecker) Check(ctx context.Context) error {
	return m.checkFunc(ctx)
}

// Test NewHealthCheckHandler
func TestNewHealthCheckHandler(t *testing.T) {
	mock := &mockChecker{
		checkFunc: func(ctx context.Context) error {
			return nil
		},
	}

	h := NewHealthCheckHandler(mock)

	if h == nil {
		t.Fatalf("expected NewHealthCheckHandler to return non-nil, got nil")
	}

	if h.ch != mock {
		t.Fatalf("expected checker to be set correctly")
	}
}

// Test Register
func TestHealthCheckHandler_Register(t *testing.T) {
	mock := &mockChecker{
		checkFunc: func(ctx context.Context) error {
			return nil
		},
	}

	h := NewHealthCheckHandler(mock)
	mux := chi.NewMux()

	h.Register(mux)

	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status code %v, got %v", http.StatusOK, recorder.Code)
	}
}

// Test Check with no errors
func TestHealthCheckHandler_Check_NoError(t *testing.T) {
	mock := &mockChecker{
		checkFunc: func(ctx context.Context) error {
			return nil
		},
	}
	h := NewHealthCheckHandler(mock)

	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	recorder := httptest.NewRecorder()

	h.Check(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status code %v, got %v", http.StatusOK, recorder.Code)
	}
}

// Test Check with checker error
func TestHealthCheckHandler_Check_WithError(t *testing.T) {
	mock := &mockChecker{
		checkFunc: func(ctx context.Context) error {
			return errors.New("internal error")
		},
	}
	h := NewHealthCheckHandler(mock)

	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	recorder := httptest.NewRecorder()

	h.Check(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status code %v, got %v", http.StatusInternalServerError, recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "internal error") {
		t.Fatalf("expected error message in response body, got %v", recorder.Body.String())
	}
}

// Test Check with context.Canceled error
func TestHealthCheckHandler_Check_ContextCanceled(t *testing.T) {
	mock := &mockChecker{
		checkFunc: func(ctx context.Context) error {
			return context.Canceled
		},
	}
	h := NewHealthCheckHandler(mock)

	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	recorder := httptest.NewRecorder()

	h.Check(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status code %v, got %v", http.StatusOK, recorder.Code)
	}
}
