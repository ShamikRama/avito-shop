package service

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/erorrs"
	"avito-shop/internal/logger"
	"avito-shop/internal/model"
	"avito-shop/internal/utils"
	"context"
	"errors"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
	"time"
)

const tokenTTL = 7 * time.Hour

//go:generate go run github.com/vektra/mockery/v2@latest --name=RepoAuthInterface
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
		return s.GenerateJwtToken(user.ID)
	case user == (domain.User{}):
		idReg, err := s.registerUser(ctx, dto)
		if err != nil {
			if errors.Is(err, erorrs.ErrUserExist) {
				s.logger.Error("service.Auth.Authorization: user exist", zap.Error(err))
				return "", erorrs.ErrUserExist
			}
			s.logger.Error("service.Auth.Authorization: authorization error", zap.Error(err))
			return "", err
		}
		return s.GenerateJwtToken(idReg)
	default:
		s.logger.Error("service.Auth.Authorization: authorization error", zap.Error(err))
		return "", err
	}

}

func (s *AuthService) authenticateUser(ctx context.Context, dto model.AuthRequestDTO) (domain.User, error) {
	username := dto.Username
	passwordHash := utils.GeneratePasswordHash(dto.Password)

	user, err := s.repo.GetUser(ctx, username, passwordHash)
	if err != nil {
		s.logger.Error("service.Auth.authenticateUser: authenticate error", zap.Error(err))
		return domain.User{}, err
	}
	if user == (domain.User{}) {
		s.logger.Info("service.Auth.authenticateUser: user not found")
		return user, nil
	}

	s.logger.Info("user authenticated successfully")

	return user, nil
}

func (s *AuthService) registerUser(ctx context.Context, dto model.AuthRequestDTO) (int, error) {
	user := utils.AuthRequestToUser(dto)

	user.PasswordHash = utils.GeneratePasswordHash(user.PasswordHash)

	id, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, erorrs.ErrUserExist) {
			s.logger.Error("service.Auth.registerUser: email already exist", zap.Error(err))
			return 0, erorrs.ErrUserExist
		}
		s.logger.Error("service.Auth.registerUser: error register", zap.Error(err))
		return 0, err
	}

	s.logger.Info("user created successfully")
	return id, nil
}

func (s *AuthService) GenerateJwtToken(userId int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		}, userId,
	})

	return token.SignedString([]byte("qrkjk#4#%35FSFJlja#4353KSFjH"))
}
