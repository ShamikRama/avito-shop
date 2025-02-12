package service

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/erorrs"
	"avito-shop/internal/logger"
	"avito-shop/internal/model"
	"avito-shop/internal/utils"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
)

type RepoAuthInterface interface {
	CreateUser(ctx context.Context, user domain.User) (int, error)
	GetUser(ctx context.Context, username string, password string) (domain.User, error)
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
	user, err := s.authenticateUser(ctx, dto)
	switch {
	case user != (domain.User{}):
		return generateJwtToken(user.ID)
	case user == (domain.User{}):
		idReg, err := s.registerUser(ctx, dto)
		if err != nil {
			if errors.Is(err, erorrs.ErrUserExist) {
				return "", fmt.Errorf("username is occupate: %w", erorrs.ErrUserExist)
			}
			return "", fmt.Errorf("registration failed: %w", err)
		}
		return generateJwtToken(idReg)
	default:
		return "", fmt.Errorf("authentication failed: %w", err)
	}
}

func (s *AuthService) authenticateUser(ctx context.Context, dto model.AuthRequestDTO) (domain.User, error) {
	username := dto.Username
	passwordHash := utils.GeneratePasswordHash(dto.Password)

	user, err := s.repo.GetUser(ctx, username, passwordHash)
	if err != nil {
		s.logger.Error("failed to get user",
			zap.String("operation", "service.GetUserId"))
	}
	if user == (domain.User{}) {
		return user, nil
	}

	return user, nil
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
		return 0, fmt.Errorf("registration failed: %w", err)
	}

	s.logger.Info("user created successfully")
	return id, nil
}
