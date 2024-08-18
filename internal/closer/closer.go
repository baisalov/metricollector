package closer

import (
	"io"
	"log/slog"
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

func (c *Closer) Close() {
	c.mx.Lock()
	defer c.mx.Unlock()

	for i := len(c.units) - 1; i >= 0; i-- {
		u := c.units[i]

		if err := u.Close(); err != nil {
			slog.Error(u.title, "error", err)
			continue
		}

		slog.Debug(u.title)
	}
}
