package checker

import (
	"context"
	"slices"
	"strings"
	"sync"
)

type Checker struct {
	mx     sync.Mutex
	checks []CheckFunc
}

func NewChecker() *Checker {
	return &Checker{}
}

type CheckFunc func(ctx context.Context) error

func (c *Checker) Register(fn CheckFunc) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.checks = append(c.checks, fn)
}

func (c *Checker) Check(ctx context.Context) error {
	c.mx.Lock()

	checks := slices.Clone(c.checks)

	c.mx.Unlock()

	var errs checkErrors

	for _, check := range checks {
		if err := check(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

type checkErrors []error

func (e checkErrors) Error() string {
	var s []string

	for _, err := range e {
		s = append(s, err.Error())
	}

	return strings.Join(s, ", ")
}

func Wrap(fn func() error) CheckFunc {
	return func(_ context.Context) error {
		return fn()
	}
}
