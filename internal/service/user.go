package service

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

type RepoUserInterface interface {
	BuyItem(ctx context.Context, userID int, item domain.Item) error
	SendCoins(ctx context.Context, fromUserID int, toUserID int, amount int) error
	GetItem(ctx context.Context, itemName string) (domain.Item, error)
	GetPurchasedItems(ctx context.Context, userID int) ([]model.ItemDTO, error)
	GetUser(ctx context.Context, userID int) (domain.User, error)
	GetUserByName(ctx context.Context, toUser string) (int, error)
	GetCoinHistory(ctx context.Context, userID int, currentUsername string) (model.CoinHistoryDTO, error)
}

type UserService struct {
	repo   RepoUserInterface
	logger logger.Logger
}

func NewUserService(repo RepoUserInterface, logger logger.Logger) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger,
	}
}

func (s *UserService) SendCoinToUser(ctx context.Context, fromUserID int, toUser string, amount int) error {
	toUserID, err := s.repo.GetUserByName(ctx, toUser)
	if err != nil {
		return fmt.Errorf("get recipient ID: %w", err)
	}

	if fromUserID == toUserID {
		return erorrs.ErrSelfTransfer
	}

	if err := s.repo.SendCoins(ctx, fromUserID, toUserID, amount); err != nil {
		switch {
		case errors.Is(err, erorrs.ErrNotFound):
			s.logger.Error("user not found", zap.String("operation", "service.SendCoin"))
			return erorrs.ErrNotFound
		case errors.Is(err, erorrs.ErrInsufficientFunds):
			s.logger.Error("insufficient funds", zap.String("operation", "service.SendCoin"))
			return erorrs.ErrInsufficientFunds
		default:
			s.logger.Error("failed to send coins", zap.Error(err))
			return fmt.Errorf("send coin: %w", err)
		}
	}

	s.logger.Info("money sent successfully", zap.Int("fromUserID", fromUserID), zap.Int("toUserID", toUserID), zap.Int("amount", amount))

	return nil
}

func (s *UserService) BuyItem(ctx context.Context, userID int, input model.BuyItemRequestDTO) error {
	item, err := s.repo.GetItem(ctx, input.Item)
	if err != nil {
		if errors.Is(err, erorrs.ErrNotFound) {
			s.logger.Error("item not found", zap.String("item", input.Item), zap.String("operation", "service.BuyItem"))
			return erorrs.ErrNotFound
		}
		s.logger.Error("failed to get item", zap.Error(err))
		return fmt.Errorf("buy item: %w", err)
	}

	if err := s.repo.BuyItem(ctx, userID, item); err != nil {
		switch {
		case errors.Is(err, erorrs.ErrNotFound):
			s.logger.Error("user not found", zap.String("operation", "service.BuyItem"))
			return erorrs.ErrNotFound
		case errors.Is(err, erorrs.ErrInsufficientFunds):
			s.logger.Error("insufficient funds", zap.String("operation", "service.BuyItem"))
			return erorrs.ErrInsufficientFunds
		default:
			s.logger.Error("failed to buy item", zap.Error(err))
			return fmt.Errorf("buy item: %w", err)
		}
	}

	s.logger.Info("item bought successfully", zap.Int("userID", userID), zap.String("item", input.Item))
	return nil
}

func (s *UserService) GetUserInfo(ctx context.Context, userID int) (model.InfoResponseDTO, error) {
	items, err := s.getUserItems(ctx, userID)
	if err != nil {
		if errors.Is(err, erorrs.ErrItemNotFound) {
			s.logger.Error("not found item for user", zap.Int("userID", userID))
			items = []model.ItemDTO{}
		} else {
			s.logger.Error("error getting items for user", zap.Int("userID", userID), zap.Error(err))
			return model.InfoResponseDTO{}, fmt.Errorf("get user items: %w", err)
		}
	}

	balance, userName, err := s.getUserNameBalance(ctx, userID)
	if err != nil {
		s.logger.Error("error getting balance for user", zap.Int("userID", userID), zap.Error(err))
		return model.InfoResponseDTO{}, fmt.Errorf("get user balance: %w", err)
	}

	coinHistory, err := s.repo.GetCoinHistory(ctx, userID, userName)
	if err != nil {
		s.logger.Error("failed to get coin history", zap.Int("userID", userID), zap.Error(err))
		return model.InfoResponseDTO{}, fmt.Errorf("get coin history: %w", err)
	}

	return model.InfoResponseDTO{
		Coins:       balance,
		Inventory:   items,
		CoinHistory: coinHistory,
	}, nil
}

func (s *UserService) getUserItems(ctx context.Context, userID int) ([]model.ItemDTO, error) {
	items, err := s.repo.GetPurchasedItems(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []model.ItemDTO{}, erorrs.ErrItemNotFound
		}
		s.logger.Error("failed to get purchased items", zap.Int("userID", userID), zap.Error(err))
		return nil, fmt.Errorf("get purchased items: %w", err)
	}
	return items, nil
}

func (s *UserService) getUserNameBalance(ctx context.Context, userID int) (int, string, error) {
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		s.logger.Error("error get balance in servicre")
		return 0, "", fmt.Errorf("get balance: %w", err)
	}

	return user.Coins, user.Username, nil
}
