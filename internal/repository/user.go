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
		r.logger.Error("sql.User.SendCoin: error begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	var currentBalance int
	err = tx.QueryRowContext(ctx,
		"SELECT balance FROM users WHERE id = $1 FOR UPDATE",
		fromUserID,
	).Scan(&currentBalance)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("sql.User.SendCoin: no rows", zap.Error(err))
			return erorrs.ErrNotFound
		}
		return err
	}

	if currentBalance < amount {
		r.logger.Error("sql.User.SendCoin: insufficient balance", zap.Error(err))
		return erorrs.ErrInsufficientFunds
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE users SET balance = balance - $1 WHERE id = $2",
		amount, fromUserID,
	)
	if err != nil {
		r.logger.Error("sql.User.SendCoin: subtract balance", zap.Error(err))
		return err
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE users SET balance = balance + $1 WHERE id = $2",
		amount, toUserID,
	)
	if err != nil {
		r.logger.Error("sql.User.SendCoin: add balance", zap.Error(err))
		return err
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO transfers (from_user_id, to_user_id, amount) VALUES ($1, $2, $3)",
		fromUserID, toUserID, amount,
	)
	if err != nil {
		r.logger.Error("sql.User.SendCoin: save transfer", zap.Error(err))
		return err
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
			r.logger.Error("sql.User.GetItem: no rows", zap.Error(err))
			return item, erorrs.ErrItemNotFound
		}
		r.logger.Error("sql.User.GetItem: error query row", zap.Error(err))
		return item, err
	}
	return item, nil
}

func (r *UserRepo) BuyItem(ctx context.Context, user domain.User, item domain.Item) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		r.logger.Error("sql.User.ByItem: error begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE users SET balance = balance - $1 WHERE id = $2`,
		item.Price, user.ID,
	)
	if err != nil {
		r.logger.Error("sql.User.ByItem: error exec", zap.Error(err))
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		r.logger.Error("sql.User.ByItem: error rows affected", zap.Error(err))
		return err
	}
	if rows == 0 {
		r.logger.Error("sql.User.ByItem: no rows", zap.Error(err))
		return erorrs.ErrNotFound
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO purchases (user_id, item_id, item_name, quantity)
        VALUES ($1, $2, $3, 1)
        ON CONFLICT (user_id, item_id) 
        DO UPDATE SET quantity = purchases.quantity + 1`,
		user.ID, item.ID, item.Name,
	)
	if err != nil {
		r.logger.Error("sql.User.ByItem: error exec", zap.Error(err))
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
		r.logger.Error("sql.User.GetPurchasedItems: error query", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var items []model.ItemDTO
	for rows.Next() {
		var item model.ItemDTO
		if err := rows.Scan(&item.Type, &item.Quantity); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				r.logger.Error("sql.User.GetPurchasedItems: no rows", zap.Error(err))
				return items, sql.ErrNoRows
			}
			r.logger.Error("sql.User.GetPurchasedItems: error scan", zap.Error(err))
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("sql.User.GetPurchasedItems: rows error", zap.Error(err))
		return nil, err
	}

	return items, nil
}

func (r *UserRepo) GetUser(ctx context.Context, userID int) (domain.User, error) {
	var user domain.User

	err := r.db.QueryRowContext(ctx, `SELECT id, username, balance FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Coins)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("sql.User.GetUser: user not found", zap.Error(err))
			return domain.User{}, erorrs.ErrNotFound
		}
		r.logger.Error("sql.User.GetUser: error query", zap.Error(err))
		return user, err
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
		r.logger.Error("sql.User.GetCoinHistory: error query", zap.Error(err))
		return model.CoinHistoryDTO{}, err
	}
	defer rows.Close()

	var received []model.TransactionHistoryDTO
	var sent []model.TransactionHistoryDTO

	for rows.Next() {
		var fromUser, toUser string
		var amount int

		if err := rows.Scan(&fromUser, &toUser, &amount); err != nil {
			r.logger.Error("sql.User.GetCoinHistory: error scan", zap.Error(err))
			return model.CoinHistoryDTO{}, err
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
		r.logger.Error("sql.User.GetCoinHistory: rows error", zap.Error(err))
		return model.CoinHistoryDTO{}, err
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
		r.logger.Error("sql.User.GetUserByName: error query", zap.Error(err))
		return 0, err
	}

	return id, nil
}
