package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	UserExists(ctx context.Context, login string) (bool, error)
	SaveUser(ctx context.Context, login, password string) error
	SaveToken(ctx context.Context, login string, token string) error
	DeleteToken(ctx context.Context, token string) (bool, error)
	GetLoginFromToken(ctx context.Context, token string) (string, error)
}

type userrepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userrepo{db: db}
}
func (u *userrepo) UserExists(ctx context.Context, login string) (bool, error) {

	query := "SELECT EXISTS(SELECT 1 FROM users WHERE login = $1)"
	var exists bool
	err := u.db.QueryRow(ctx, query, login).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (u *userrepo) SaveUser(ctx context.Context, login, password string) error {
	query := `insert into  users (login,password) values($1,$2)`

	res, err := u.db.Exec(ctx, query, login, password)
	rowAff := res.RowsAffected()
	fmt.Println("from save user", rowAff, err)
	if err != nil {
		return err
	}
	return nil

}
func (u *userrepo) SaveToken(ctx context.Context, login string, token string) error {
	query := `
		UPDATE users
		SET token = $2
		WHERE login = $1;
	`
	res, err := u.db.Exec(ctx, query, login, token)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	rowsAffected := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("no rows were affected")
	}

	return nil
}
func (u *userrepo) DeleteToken(ctx context.Context, token string) (bool, error) {

	query := `UPDATE users SET token = NULL WHERE token = $1`

	res, err := u.db.Exec(ctx, query, token)
	if err != nil {
		return false, fmt.Errorf("failed to execute query: %w", err)
	}

	rowsAffected := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return false, nil
	}

	return true, nil
}
func (u *userrepo) GetLoginFromToken(ctx context.Context, token string) (string, error) {

	query := `SELECT login FROM users WHERE token = $1`
	row := u.db.QueryRow(ctx, query, token)
	var login string
	if err := row.Scan(&login); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return login, nil
}
