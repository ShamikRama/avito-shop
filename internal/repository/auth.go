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

func (r *AuthRepo) GetUserId(ctx context.Context, username string, password string) (int, error) {
	var id int

	query := `SELECT id FROM users WHERE username = $1 AND password_hash = $2`

	err := r.db.QueryRowContext(ctx, query, username).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("No rows in sql for id")
			return 0, sql.ErrNoRows
		}
	}

	return id, nil
}

func (r *AuthRepo) CreateUser(ctx context.Context, user domain.User) (int, error) {
	query := `INSERT INTO users (username, password_hash, balance) 
              VALUES ($1, $2, $3) 
              RETURNING id`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		user.Username,
		user.PasswordHash,
		user.Coins,
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
