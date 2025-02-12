package service

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/erorrs"
	"avito-shop/internal/logger"
	"avito-shop/internal/model"
	"avito-shop/internal/utils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go.uber.org/zap"
)

type RepoAuthInterface interface {
	CreateUser(ctx context.Context, user domain.User) (int, error)
	GetUserId(ctx context.Context, username string, password string) (int, error)
}

type AuthService struct {
	repo   RepoAuthInterface
	logger logger.Logger
}

func NewAuthService(repo RepoAuthInterface, logger logger.Logger) *AuthService {
	return &AuthService{
		repo:   repo,
		logger: logger,
	}
}

func (s *AuthService) Authorization(ctx context.Context, dto model.AuthRequestDTO) (string, error) {
	idLog, err := s.authenticateUser(ctx, dto)
	switch {
	case errors.Is(err, nil):
		return generateJwtToken(idLog)
	case errors.Is(err, erorrs.ErrNotFound):
		idReg, err := s.registerUser(ctx, dto)
		if err != nil {
			return "", fmt.Errorf("registration failed: %w", err)
		}
		return generateJwtToken(idReg)

	default:
		return "", fmt.Errorf("authentication failed: %w", err)
	}
}

func (s *AuthService) authenticateUser(ctx context.Context, dto model.AuthRequestDTO) (int, error) {
	username := dto.Username
	passwordHash := utils.GeneratePasswordHash(dto.Password)

	id, err := s.repo.GetUserId(ctx, username, passwordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, erorrs.ErrNotFound
		}
		s.logger.Error("failed to get user",
			zap.String("operation", "service.GetUserId"))
		return 0, fmt.Errorf("failed to get user: %w", err)
	}
	return id, nil
}

func (s *AuthService) registerUser(ctx context.Context, dto model.AuthRequestDTO) (int, error) {
	user := utils.AuthRequestToUser(dto)

	user.PasswordHash = utils.GeneratePasswordHash(user.PasswordHash)

	id, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, erorrs.ErrUserExist) {
			s.logger.Error("email already exist",
				zap.String("operation", "service.CreateUser"))
			return 0, erorrs.ErrUserExist
		}
		s.logger.Error("error email create user",
			zap.String("operation", "service.CreateUser"))
		return 0, err
	}

	s.logger.Info("user created successfully")
	return id, nil
}
