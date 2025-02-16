//go:build unit
// +build unit

package service

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/erorrs"
	mocks2 "avito-shop/internal/logger/mocks"
	"avito-shop/internal/model"
	"avito-shop/internal/service/mocks"
	"avito-shop/internal/utils"
	"context"
	"errors"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestAuthService_Authorization(t *testing.T) {
	mockRepo := mocks.NewRepoAuthInterface(t)
	mockLogger := mocks2.NewLogger(t)
	authService := NewAuthService(mockRepo, mockLogger)

	tests := []struct {
		name          string
		dto           model.AuthRequestDTO
		mockRepo      func()
		mockLogger    func()
		expectToken   bool
		expectedError error
	}{
		{
			name: "successful authentication",
			dto:  model.AuthRequestDTO{Username: "user", Password: "pass"},
			mockRepo: func() {
				mockRepo.On("GetUser", mock.Anything, "user", utils.GeneratePasswordHash("pass")).
					Return(domain.User{ID: 1}, nil)
			},
			mockLogger: func() {
				mockLogger.On("Info", mock.Anything, mock.Anything).Once()
			},
			expectToken:   true,
			expectedError: nil,
		},
		{
			name: "User not found, successful registration",
			dto:  model.AuthRequestDTO{Username: "new_user", Password: "password"},
			mockRepo: func() {
				mockRepo.On("GetUser", mock.Anything, "new_user", utils.GeneratePasswordHash("password")).
					Return(domain.User{}, nil)
				mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(u domain.User) bool {
					return u.Username == "new_user" &&
						u.PasswordHash == utils.GeneratePasswordHash("password")
				})).Return(2, nil)
			},
			mockLogger: func() {
				mockLogger.On("Info", mock.Anything)
			},
			expectToken:   true,
			expectedError: nil,
		},
		{
			name: "user exists during registration",
			dto:  model.AuthRequestDTO{Username: "exists", Password: "pass"},
			mockRepo: func() {
				mockRepo.On("GetUser", mock.Anything, "exists", mock.Anything).
					Return(domain.User{}, nil)
				mockRepo.On("CreateUser", mock.Anything, mock.Anything).
					Return(0, erorrs.ErrUserExist)
			},
			mockLogger: func() { // Добавляем ожидание вызова Info
				mockLogger.On("Info", mock.Anything, mock.Anything).Once()
				mockLogger.On("Error", mock.Anything, mock.Anything).Twice()
			},
			expectToken:   false,
			expectedError: erorrs.ErrUserExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockLogger.ExpectedCalls = nil
			tt.mockRepo()
			tt.mockLogger()

			token, err := authService.Authorization(context.Background(), tt.dto)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
			} else {
				assert.NoError(t, err)
			}

			if tt.expectToken {
				assert.NotEmpty(t, token)
			} else {
				assert.Empty(t, token)
			}

			mockRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestAuthService_AuthenticateUser(t *testing.T) {
	mockRepo := mocks.NewRepoAuthInterface(t)
	mockLogger := mocks2.NewLogger(t)
	authService := NewAuthService(mockRepo, mockLogger)

	tests := []struct {
		name          string
		dto           model.AuthRequestDTO
		mockRepo      func()
		mockLogger    func()
		expectedUser  domain.User
		expectedError error
	}{
		{
			name: "user found",
			dto:  model.AuthRequestDTO{Username: "user", Password: "pass"},
			mockLogger: func() {
				mockLogger.On("Info", mock.Anything).Once()
			},
			mockRepo: func() {
				mockRepo.On("GetUser", mock.Anything, "user", utils.GeneratePasswordHash("pass")).
					Return(domain.User{ID: 1}, nil)
			},
			expectedUser:  domain.User{ID: 1},
			expectedError: nil,
		},
		{
			name: "user not found",
			dto:  model.AuthRequestDTO{Username: "nonexistent", Password: "pass"},
			mockLogger: func() {
				mockLogger.On("Info", mock.Anything).Once()
			},
			mockRepo: func() {
				mockRepo.On("GetUser", mock.Anything, "nonexistent", mock.Anything).
					Return(domain.User{}, nil)
			},
			expectedUser:  domain.User{},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockRepo()
			tt.mockLogger()

			user, err := authService.authenticateUser(context.Background(), tt.dto)

			assert.Equal(t, tt.expectedUser, user)
			assert.Equal(t, tt.expectedError, err)
			mockRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestAuthService_RegisterUser(t *testing.T) {
	mockRepo := mocks.NewRepoAuthInterface(t)
	mockLogger := mocks2.NewLogger(t)
	authService := NewAuthService(mockRepo, mockLogger)

	tests := []struct {
		name          string
		dto           model.AuthRequestDTO
		mockRepo      func()
		mockLogger    func()
		expectedID    int
		expectedError error
	}{
		{
			name: "successful registration",
			dto:  model.AuthRequestDTO{Username: "new", Password: "pass"},
			mockRepo: func() {
				mockRepo.On("CreateUser", mock.Anything, utils.AuthRequestToUser(model.AuthRequestDTO{Username: "new", Password: utils.GeneratePasswordHash("pass")})).
					Return(1, nil)
			},
			mockLogger: func() {
				mockLogger.On("Info", mock.Anything)
			},
			expectedID:    1,
			expectedError: nil,
		},
		{
			name: "user already exists",
			dto:  model.AuthRequestDTO{Username: "exists", Password: "pass"},
			mockRepo: func() {
				mockRepo.On("CreateUser", mock.Anything, mock.Anything).
					Return(0, erorrs.ErrUserExist)
			},
			mockLogger: func() {
				mockLogger.On("Error", mock.Anything, mock.Anything)
			},
			expectedID:    0,
			expectedError: erorrs.ErrUserExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockRepo()
			tt.mockLogger()

			id, err := authService.registerUser(context.Background(), tt.dto)

			assert.Equal(t, tt.expectedID, id)
			assert.Equal(t, tt.expectedError, err)
			mockRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestAuthService_GenerateJwtToken(t *testing.T) {
	authService := NewAuthService(nil, nil)

	tests := []struct {
		name        string
		userID      int
		checkClaims func(*testing.T, string)
	}{
		{
			name:   "valid token generation",
			userID: 123,
			checkClaims: func(t *testing.T, tokenString string) {
				token, err := jwt.ParseWithClaims(tokenString, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte("qrkjk#4#%35FSFJlja#4353KSFjH"), nil
				})

				assert.NoError(t, err)
				assert.True(t, token.Valid)

				if claims, ok := token.Claims.(*tokenClaims); ok {
					assert.Equal(t, 123, claims.UserId)
					assert.True(t, claims.ExpiresAt > time.Now().Unix())
				} else {
					t.Fatal("invalid claims type")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := authService.GenerateJwtToken(tt.userID)
			assert.NoError(t, err)
			tt.checkClaims(t, token)
		})
	}
}
