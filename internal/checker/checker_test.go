package checker

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// Helper function to create a CheckFunc that returns a specific error.
func createCheckFunc(err error) CheckFunc {
	return func(ctx context.Context) error {
		return err
	}
}

// Test NewChecker
func TestNewChecker(t *testing.T) {
	c := NewChecker()
	if c == nil {
		t.Fatalf("expected NewChecker to return non-nil, got nil")
	}
	if len(c.checks) != 0 {
		t.Fatalf("expected NewChecker to initialize an empty checks slice, got %v", len(c.checks))
	}
}

// Test Register
func TestChecker_Register(t *testing.T) {
	c := NewChecker()
	fn := createCheckFunc(nil)
	c.Register(fn)

	if len(c.checks) != 1 {
		t.Fatalf("expected 1 check function registered, got %d", len(c.checks))
	}

	c.Register(fn)
	if len(c.checks) != 2 {
		t.Fatalf("expected 2 check functions registered, got %d", len(c.checks))
	}
}

// Test Check with no errors
func TestChecker_Check_NoErrors(t *testing.T) {
	c := NewChecker()
	ctx := context.Background()

	// Test with no checks registered
	if err := c.Check(ctx); err != nil {
		t.Fatalf("expected no error when no checks are registered, got %v", err)
	}

	// Test with one successful check
	c.Register(createCheckFunc(nil))
	if err := c.Check(ctx); err != nil {
		t.Fatalf("expected no error when all checks pass, got %v", err)
	}

	// Test with multiple successful checks
	c.Register(createCheckFunc(nil))
	c.Register(createCheckFunc(nil))
	if err := c.Check(ctx); err != nil {
		t.Fatalf("expected no error when all checks pass, got %v", err)
	}
}

// Test Check with errors
func TestChecker_Check_WithErrors(t *testing.T) {
	c := NewChecker()
	ctx := context.Background()

	// Test with one failing check
	c.Register(createCheckFunc(errors.New("check failed")))
	err := c.Check(ctx)
	if err == nil {
		t.Fatal("expected an error when a check fails, got nil")
	}

	if !strings.Contains(err.Error(), "check failed") {
		t.Fatalf("expected error message to contain 'check failed', got %v", err.Error())
	}

	// Test with multiple checks, some failing
	c = NewChecker() // reset checker
	c.Register(createCheckFunc(errors.New("first failure")))
	c.Register(createCheckFunc(nil))
	c.Register(createCheckFunc(errors.New("second failure")))

	err = c.Check(ctx)
	if err == nil {
		t.Fatal("expected an error when some checks fail, got nil")
	}

	expectedErrs := []string{"first failure", "second failure"}
	for _, expectedErr := range expectedErrs {
		if !strings.Contains(err.Error(), expectedErr) {
			t.Fatalf("expected error message to contain '%v', got '%v'", expectedErr, err.Error())
		}
	}
}

// Test Wrap function
func TestWrap(t *testing.T) {
	fn := func() error {
		return errors.New("wrapped error")
	}

	wrappedFn := Wrap(fn)
	err := wrappedFn(context.Background())

	if err == nil {
		t.Fatal("expected error from wrapped function, got nil")
	}

	expectedErr := "wrapped error"
	if err.Error() != expectedErr {
		t.Fatalf("expected error message to be '%v', got '%v'", expectedErr, err.Error())
	}
}
