package closer

import (
	"errors"
	"fmt"
	"io"
	"sync"
)

type Closer struct {
	mx    sync.Mutex
	units []unit
}

func NewCloser() *Closer {
	return &Closer{}
}

type unit struct {
	title string
	io.Closer
}

func (c *Closer) Register(title string, cl io.Closer) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.units = append(c.units, unit{title: title, Closer: cl})
}

func (c *Closer) Close() error {
	c.mx.Lock()
	defer c.mx.Unlock()

	var errs []error

	for i := len(c.units) - 1; i >= 0; i-- {
		u := c.units[i]

		if err := u.Close(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", u.title, err))
			continue
		}
	}

	return errors.Join(errs...)
}
