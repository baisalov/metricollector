package postgres

import (
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

func retry(fn func() error) error {

	i, n := 1, 3

	for {

		err := fn()

		var pgErr *pgconn.PgError

		if err == nil || !errors.As(err, &pgErr) || !pgerrcode.IsConnectionException(pgErr.Code) || i > n {
			return err
		}

		time.Sleep(time.Duration(i+(i-1)) * time.Second)

		i++
	}
}
