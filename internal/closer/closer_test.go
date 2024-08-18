package closer

import (
	"errors"
	"strings"
	"testing"
)

// Mock implementation of io.Closer for testing purposes
type mockCloser struct {
	closeFunc func() error
}

func (m *mockCloser) Close() error {
	return m.closeFunc()
}

// Test NewCloser
func TestNewCloser(t *testing.T) {
	c := NewCloser()
	if c == nil {
		t.Fatalf("expected NewCloser to return non-nil, got nil")
	}
	if len(c.units) != 0 {
		t.Fatalf("expected NewCloser to initialize an empty units slice, got %v", len(c.units))
	}
}

// Test Register
func TestCloser_Register(t *testing.T) {
	c := NewCloser()
	mock := &mockCloser{}

	c.Register("unit1", mock)
	if len(c.units) != 1 {
		t.Fatalf("expected 1 unit registered, got %d", len(c.units))
	}

	c.Register("unit2", mock)
	if len(c.units) != 2 {
		t.Fatalf("expected 2 units registered, got %d", len(c.units))
	}

	if c.units[0].title != "unit1" || c.units[1].title != "unit2" {
		t.Fatalf("unexpected unit titles: %v, %v", c.units[0].title, c.units[1].title)
	}
}

// Test Close with all successful closures
func TestCloser_Close_AllSuccess(t *testing.T) {
	c := NewCloser()

	mock1 := &mockCloser{closeFunc: func() error { return nil }}
	mock2 := &mockCloser{closeFunc: func() error { return nil }}

	c.Register("unit1", mock1)
	c.Register("unit2", mock2)

	err := c.Close()
	if err != nil {
		t.Fatalf("expected no error when all units close successfully, got %v", err)
	}
}

// Test Close with errors
func TestCloser_Close_WithErrors(t *testing.T) {
	c := NewCloser()

	mock1 := &mockCloser{closeFunc: func() error { return errors.New("error in unit1") }}
	mock2 := &mockCloser{closeFunc: func() error { return nil }}
	mock3 := &mockCloser{closeFunc: func() error { return errors.New("error in unit3") }}

	c.Register("unit1", mock1)
	c.Register("unit2", mock2)
	c.Register("unit3", mock3)

	err := c.Close()

	if err == nil {
		t.Fatal("expected an error when some units fail to close, got nil")
	}

	errMsg := err.Error()
	expectedErrors := []string{"unit3: error in unit3", "unit1: error in unit1"}

	for _, expected := range expectedErrors {
		if !strings.Contains(errMsg, expected) {
			t.Errorf("expected error message to contain '%s', got '%s'", expected, errMsg)
		}
	}

	if strings.Contains(errMsg, "unit2") {
		t.Errorf("expected error message not to contain 'unit2', got '%s'", errMsg)
	}
}

// Test Close with no registered units
func TestCloser_Close_NoUnits(t *testing.T) {
	c := NewCloser()

	err := c.Close()
	if err != nil {
		t.Fatalf("expected no error when no units are registered, got %v", err)
	}
}
