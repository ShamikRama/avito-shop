package repository

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/erorrs"
	"avito-shop/internal/logger"
	"avito-shop/internal/model"
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

func (r *UserRepo) GetPurchasedItems(ctx context.Context, userID int) ([]model.ItemDTO, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT item_name, quantity FROM purchases WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get purchased items: %w", err)
	}
	defer rows.Close()

	var items []model.ItemDTO
	for rows.Next() {
		var item model.ItemDTO
		if err := rows.Scan(&item.Type, &item.Quantity); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				r.logger.Error("no rows")
				return items, fmt.Errorf("sql no rows: %w", sql.ErrNoRows)
			}
			return nil, fmt.Errorf("scan purchased item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return items, nil
}

func (r *UserRepo) GetUser(ctx context.Context, userID int) (domain.User, error) {
	var user domain.User

	err := r.db.QueryRowContext(ctx, `SELECT id, username, balance FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Coins)

	if err != nil {
		return user, fmt.Errorf("find balance: %w", err)
	}

	return user, nil
}

func (r *UserRepo) GetCoinHistory(ctx context.Context, userID int, currentUsername string) (model.CoinHistoryDTO, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT
            fu.username AS from_user,
            tu.username AS to_user,
            t.amount
         FROM transfers t
         JOIN users fu ON t.from_user_id = fu.id
         JOIN users tu ON t.to_user_id = tu.id
         WHERE t.from_user_id = $1 OR t.to_user_id = $1`,
		userID,
	)
	if err != nil {
		return model.CoinHistoryDTO{}, fmt.Errorf("get coin history: %w", err)
	}
	defer rows.Close()

	var received []model.TransactionHistoryDTO
	var sent []model.TransactionHistoryDTO

	for rows.Next() {
		var fromUser, toUser string
		var amount int

		if err := rows.Scan(&fromUser, &toUser, &amount); err != nil {
			return model.CoinHistoryDTO{}, fmt.Errorf("scan coin history: %w", err)
		}

		if toUser == currentUsername {
			received = append(received, model.TransactionHistoryDTO{
				FromUser: fromUser,
				Amount:   amount,
			})
		} else if fromUser == currentUsername {
			sent = append(sent, model.TransactionHistoryDTO{
				ToUser: toUser,
				Amount: amount,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return model.CoinHistoryDTO{}, fmt.Errorf("rows error: %w", err)
	}

	return model.CoinHistoryDTO{
		Received: received,
		Sent:     sent,
	}, nil
}

func (r *UserRepo) GetUserByName(ctx context.Context, toUser string) (int, error) {
	var id int

	err := r.db.QueryRowContext(ctx, `SELECT id FROM users WHERE username = $1`,
		toUser,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("get user: %w", err)
	}

	return id, nil
}
