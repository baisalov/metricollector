package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type UserStorage struct {
	db *sql.DB
}

type dbReader interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type dbManager interface {
	dbReader
	dbWriter

}

type dbWriter interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}


func (s UserStorage) doTX(ctx context.Context, txFunc func(ctx context.Context, db dbManager) error) (err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, fmt.Errorf("rollback: %w", tx.Rollback()))
		}
	}()

	err = txFunc(ctx, tx)

	if err != nil {
		return errors.Join(err, tx.Rollback())
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

type User struct{
	ID string
}


// GetUerByEmail это бубличный API компоннета, к нему могут обращаться снаружи.
// Его задача просто быть прокси и добавить в вызову источник данных.
// Так же в нем можно сделать дополнительное логирование, метрики и трейсинг.
func (s UserStorage) GetUerByEmail(ctx context.Context, email string) (User, error) {
	// тут транзакция не нужна, можно использовать sql.DB
	return s.getUserByEmail(ctx, s.db, email)
}

func (s UserStorage) getUserByEmail(ctx context.Context, q dbReader, email string) (User, error) {
	row := q.QueryRowContext(
		ctx,
		"SELECT FROM users WHERE email = $1",
		email)

	err := row.Scan()
	if err != nil {
		return User{}, fmt.Errorf("query user: %w", err)
	}

	return User{}, nil
}

func (s UserStorage) createUser(ctx context.Context, writer dbWriter, email string) error {
	_, err := writer.ExecContext(ctx, "INSERT INTO users (email) VALUES ($1)", email)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (s UserStorage) createBalance(ctx context.Context, writer dbWriter, userID string) error {
	_, err := writer.ExecContext(ctx, "INSERT INTO balances (user_id, sum) VALUES ($1, 0)", userID)
	if err != nil {
		return err
	}

	return nil
}

// нужно сделать две сущности - юзер и баланс юзера, пусть это будут две разных таблицы.
func (s UserStorage) CreateUserAndDefaultBalance(ctx context.Context, email string) error {
	err := s.doTX(ctx, func(ctx context.Context, db dbManager) error {
		// сначала нужно создать пользователя
		err := s.createUser(ctx, db, email)
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		// теперь получить его ID в рамках текущий транзакции
		user, err := s.getUserByEmail(ctx, db, email)
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}

		// теперь создать новую запись в таблице балансов
		err = s.createBalance(ctx, db, user.ID)
		if err !=  {
			return fmt.Errorf("create balance: %w", err)
		}

		return nil // тут обертка сама сделает Commit
	})

	if err != nil {
		return err
	}
}

