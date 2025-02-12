package service

import (
	"avito-shop/internal/erorrs"
	"avito-shop/internal/logger"
	"avito-shop/internal/model"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
)

type RepoUserInterface interface {
	BuyItem(ctx context.Context, userID int, itemID int, price int) error
	SendCoins(ctx context.Context, fromUserID int, toUserID int, amount int) error
	GetUserID(ctx context.Context, username string) (int, error)
	GetItemID(ctx context.Context, itemName string) (int, int, error)
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

func (s *UserService) SendCoinToUser(ctx context.Context, fromUserID int, toUserName string, amount int) error {
	toUserID, err := s.repo.GetUserID(ctx, toUserName)
	if err != nil {
		return fmt.Errorf("get recipient ID: %w", err)
	}

	if fromUserID == toUserID {
		return erorrs.ErrSelfTransfer
	}

	err = s.repo.SendCoins(ctx, fromUserID, toUserID, amount)
	if err != nil {
		if errors.Is(err, erorrs.ErrNotFound) {
			s.logger.Info("user not found",
				zap.String("operation", "service.SendCoin"))
			return erorrs.ErrNotFound
		} else if errors.Is(err, erorrs.ErrInsufficientFunds) {
			s.logger.Info("balance is wrong",
				zap.String("operation", "service.SendCoin"))
			return erorrs.ErrInsufficientFunds
		}

		return fmt.Errorf("send coin: %w", err)
	}

	s.logger.Info("Money sent successfully")

	return nil
}

func (s *UserService) BuyItem(ctx context.Context, userID int, input model.BuyItemRequestDTO) error {
	itemID, priceItem, err := s.repo.GetItemID(ctx, input.Item)
	if err != nil {
		if errors.Is(err, erorrs.ErrNotFound) {
			s.logger.Error("item not found",
				zap.String("operation", "service.BuyItem"))
			return erorrs.ErrNotFound
		}
		return fmt.Errorf("buy item: %w", err)
	}

	err = s.repo.BuyItem(ctx, userID, itemID, priceItem)
	if err != nil {
		if errors.Is(err, erorrs.ErrNotFound) {
			s.logger.Info("user not found",
				zap.String("operation", "service.BuyItem"))
			return erorrs.ErrNotFound
		} else if errors.Is(err, erorrs.ErrInsufficientFunds) {
			s.logger.Info("balance is wrong",
				zap.String("operation", "service.BuyItem"))
			return erorrs.ErrInsufficientFunds
		}

		return fmt.Errorf("send coin: %w", err)
	}

	s.logger.Info("Buy item successfully")

	return nil
}
