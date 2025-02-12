package repository

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/erorrs"
	"avito-shop/internal/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

type AuthRepo struct {
	db     *sql.DB
	logger logger.Logger
}

func NewAuthRepo(db *sql.DB, logger logger.Logger) *AuthRepo {
	return &AuthRepo{
		db:     db,
		logger: logger,
	}
}

func (r *AuthRepo) GetUser(ctx context.Context, username string, password string) (domain.User, error) {
	var user domain.User

	query := `
        SELECT id, username, password_hash, balance 
        FROM users 
        WHERE username = $1 AND password_hash = $2
    `

	err := r.db.QueryRowContext(ctx, query, username, password).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Coins,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("No rows in sql for user", zap.Error(err))
			return domain.User{}, nil
		}
		r.logger.Error("Database error: get user", zap.Error(err))
		return domain.User{}, fmt.Errorf("database error: %w", err)
	}

	return user, nil
}

func (r *AuthRepo) CreateUser(ctx context.Context, user domain.User) (int, error) {
	query := `INSERT INTO users (username, password_hash, balance) 
              VALUES ($1, $2, $3) 
              RETURNING id`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		user.Username,
		user.PasswordHash,
		1000,
	).Scan(&id)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return 0, erorrs.ErrUserExist
		}
		r.logger.Error("Database error: create user", zap.Error(err))
		return 0, fmt.Errorf("database error: %w", err)
	}

	return id, nil
}
