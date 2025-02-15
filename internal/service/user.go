package service

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/erorrs"
	"avito-shop/internal/logger"
	"avito-shop/internal/model"
	"context"
	"database/sql"
	"errors"
	"go.uber.org/zap"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=RepoUserInterface
type RepoUserInterface interface {
	BuyItem(ctx context.Context, user domain.User, item domain.Item) error // в процессе
	SendCoins(ctx context.Context, fromUserID int, toUserID int, amount int) error
	GetItem(ctx context.Context, itemName string) (domain.Item, error) // +
	GetPurchasedItems(ctx context.Context, userID int) ([]model.ItemDTO, error)
	GetUser(ctx context.Context, userID int) (domain.User, error) // +
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
		s.logger.Error("service.User.SendCoinToUser: error getting user", zap.Error(err))
		return err
	}

	if fromUserID == toUserID {
		s.logger.Error("service.User.SendCoinToUser: ", zap.Error(err))
		return erorrs.ErrSelfTransfer
	}

	if err := s.repo.SendCoins(ctx, fromUserID, toUserID, amount); err != nil {
		switch {
		case errors.Is(err, erorrs.ErrNotFound):
			s.logger.Error("service.User.SendCoinToUser: user not found", zap.Error(err))
			return erorrs.ErrNotFound
		case errors.Is(err, erorrs.ErrInsufficientFunds):
			s.logger.Error("service.User.SendCoinToUser: money not enough", zap.Error(err))
			return erorrs.ErrInsufficientFunds
		default:
			s.logger.Error("service.User.SendCoinToUser: error sending coin", zap.Error(err))
			return err
		}
	}

	s.logger.Info("money sent successfully", zap.Int("fromUserID", fromUserID), zap.Int("toUserID", toUserID), zap.Int("amount", amount))

	return nil
}

func (s *UserService) BuyItem(ctx context.Context, userID int, input model.BuyItemRequestDTO) error {
	item, err := s.repo.GetItem(ctx, input.Item)
	if err != nil {
		if errors.Is(err, erorrs.ErrItemNotFound) {
			s.logger.Error("service.User.BuyItem: item not found", zap.Error(err))
			return erorrs.ErrItemNotFound
		}
		s.logger.Error("service.User.BuyItem: error getting item", zap.Error(err))
		return err
	}

	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, erorrs.ErrNotFound) {
			s.logger.Error("service.User.BuyItem: user not found", zap.Error(err))
			return erorrs.ErrNotFound
		}
		s.logger.Error("service.User.BuyItem: error getting user", zap.Error(err))
		return err
	}

	if user.Coins < item.Price {
		s.logger.Error("service.User.ByItem: insufficient funds", zap.Error(err))
		return erorrs.ErrInsufficientFunds
	}

	err = s.repo.BuyItem(ctx, user, item)
	if err != nil {
		if errors.Is(err, erorrs.ErrNotFound) {
			s.logger.Error("service.User.ByItem: rows not found", zap.Error(err))
			return erorrs.ErrNotFound
		}
		s.logger.Error("service.User.ByItem: error buy", zap.Error(err))
		return err
	}

	s.logger.Info("item bought successfully", zap.Int("userID", userID))
	return nil
}

func (s *UserService) GetUserInfo(ctx context.Context, userID int) (model.InfoResponseDTO, error) {
	items, err := s.getUserItems(ctx, userID)
	if err != nil {
		if errors.Is(err, erorrs.ErrItemNotFound) {
			s.logger.Error("service.User.GetUserInfo: items is null", zap.Error(err))
			items = []model.ItemDTO{}
		} else {
			s.logger.Error("service.User.GetUserInfo: error getting info", zap.Error(err))
			return model.InfoResponseDTO{}, err
		}
	}

	balance, userName, err := s.getUserNameBalance(ctx, userID)
	if err != nil {
		s.logger.Error("service.User.GetUserInfo: error getting user name and balance", zap.Error(err))
		return model.InfoResponseDTO{}, err
	}

	coinHistory, err := s.repo.GetCoinHistory(ctx, userID, userName)
	if err != nil {
		s.logger.Error("service.User.GetUserInfo: error get coin history", zap.Error(err))
		return model.InfoResponseDTO{}, err
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
			s.logger.Error("service.User.getUserItems: no items", zap.Error(err))
			return []model.ItemDTO{}, erorrs.ErrItemNotFound
		}
		s.logger.Error("service.User.getUserItems: error get user items", zap.Error(err))
		return nil, err
	}
	return items, nil
}

func (s *UserService) getUserNameBalance(ctx context.Context, userID int) (int, string, error) {
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		s.logger.Error("service.User.getUserNameBalance: error", zap.Error(err))
		return 0, "", err
	}

	return user.Coins, user.Username, nil
}
