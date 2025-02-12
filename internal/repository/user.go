package repository

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/erorrs"
	"avito-shop/internal/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go.uber.org/zap"
)

type UserRepo struct {
	db     *sql.DB
	logger logger.Logger
}

func NewUserRepo(db *sql.DB, logger logger.Logger) *UserRepo {
	return &UserRepo{
		db:     db,
		logger: logger,
	}
}

func (r *UserRepo) SendCoins(ctx context.Context, fromUserID int, toUserID int, amount int) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var currentBalance int
	err = tx.QueryRowContext(ctx,
		"SELECT balance FROM users WHERE id = $1 FOR UPDATE",
		fromUserID,
	).Scan(&currentBalance)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("user not found",
				zap.String("operation", "sql.SendCoin"))
			return erorrs.ErrNotFound
		}
		return fmt.Errorf("get sender balance: %w", err)
	}

	if currentBalance < amount {
		r.logger.Error("insufficient balance",
			zap.String("operation", "sql.SendCoin"))
		return erorrs.ErrInsufficientFunds
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE users SET balance = balance - $1 WHERE id = $2",
		amount, fromUserID,
	)
	if err != nil {
		return fmt.Errorf("subtract balance: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE users SET balance = balance + $1 WHERE id = $2",
		amount, toUserID,
	)
	if err != nil {
		return fmt.Errorf("add balance: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO transfers (from_user_id, to_user_id, amount) VALUES ($1, $2, $3)",
		fromUserID, toUserID, amount,
	)
	if err != nil {
		return fmt.Errorf("save transfer: %w", err)
	}

	return tx.Commit()
}

func (r *UserRepo) GetUserID(ctx context.Context, username string) (int, error) {
	var userID int
	err := r.db.QueryRowContext(ctx,
		"SELECT id FROM users WHERE username = $1",
		username,
	).Scan(&userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, erorrs.ErrNotFound
		}
		return 0, fmt.Errorf("get user ID: %w", err)
	}
	return userID, nil
}

func (r *UserRepo) GetItem(ctx context.Context, itemName string) (domain.Item, error) {
	var item domain.Item

	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, price FROM items WHERE name = $1`,
		itemName,
	).Scan(&item.ID, &item.Name, &item.Price)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return item, erorrs.ErrNotFound
		}
		return item, fmt.Errorf("get item: %w", err)
	}
	return item, nil
}

func (r *UserRepo) BuyItem(ctx context.Context, userID int, item domain.Item) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var balance int
	err = tx.QueryRowContext(ctx,
		`SELECT balance FROM users WHERE id = $1 FOR UPDATE`,
		userID,
	).Scan(&balance)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return erorrs.ErrNotFound
		}
		return fmt.Errorf("get balance: %w", err)
	}

	if balance < item.Price {
		return erorrs.ErrInsufficientFunds
	}

	res, err := tx.ExecContext(ctx,
		`UPDATE users SET balance = balance - $1 WHERE id = $2`,
		item.Price, userID,
	)
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("check balance update: %w", err)
	}
	if rows == 0 {
		return erorrs.ErrNotFound
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO purchases (user_id, item_id, item_name, quantity)
        VALUES ($1, $2, $3, 1)
        ON CONFLICT (user_id, item_id) 
        DO UPDATE SET quantity = purchases.quantity + 1`,
		userID, item.ID, item.Name,
	)
	if err != nil {
		return fmt.Errorf("update purchases: %w", err)
	}

	return tx.Commit()
}
