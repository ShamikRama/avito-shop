package tests

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/erorrs"
	mocks2 "avito-shop/internal/logger/mocks"
	"avito-shop/internal/model"
	"avito-shop/internal/service"
	"avito-shop/internal/service/mocks"
	"avito-shop/internal/utils"
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestAuthorization(t *testing.T) {
	// Создаем мок репозитория
	mockRepoAuth := mocks.NewRepoAuthInterface(t)

	// Создаем мок логгера
	mockLogger := mocks2.NewLogger(t)

	// Создаем сервис с моком репозитория и логгера
	authService := service.NewAuthService(mockRepoAuth, mockLogger)

	tests := []struct {
		name          string
		dto           model.AuthRequestDTO
		mockSetup     func(*mocks.RepoAuthInterface)
		loggerSetup   func(*mocks2.Logger) // Добавляем настройку для логгера
		expectToken   bool                 // Ожидаем ли мы токен (true, если токен должен быть не пустым)
		expectedError error
	}{
		{
			name: "Successful authentication",
			dto:  model.AuthRequestDTO{Username: "existing_user", Password: "password"},
			mockSetup: func(mockRepo *mocks.RepoAuthInterface) {
				// Настраиваем мок для успешного вызова GetUser
				mockRepo.On("GetUser", mock.Anything, "existing_user", utils.GeneratePasswordHash("password")).
					Return(domain.User{ID: 1, Username: "existing_user"}, nil)
			},
			loggerSetup: func(mockLogger *mocks2.Logger) {
				// Настраиваем мок для вызова Info с одним аргументом
				mockLogger.On("Info", mock.AnythingOfType("string"), mock.AnythingOfType("zapcore.Field"), mock.AnythingOfType("zapcore.Field")).Return()
			},
			expectToken:   true, // Ожидаем не пустой токен
			expectedError: nil,
		},
		{
			name: "User not found, successful registration",
			dto:  model.AuthRequestDTO{Username: "new_user", Password: "password"},
			mockSetup: func(mockRepo *mocks.RepoAuthInterface) {
				// Настраиваем мок для возврата пустого пользователя
				mockRepo.On("GetUser", mock.Anything, "new_user", utils.GeneratePasswordHash("password")).
					Return(domain.User{}, nil)
				// Настраиваем мок для успешного вызова CreateUser
				mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("domain.User")).
					Return(2, nil)
			},
			loggerSetup: func(mockLogger *mocks2.Logger) {
				// Настраиваем мок для вызова Info с одним аргументом
				mockLogger.On("Info", "user created successfully").Return()
			},
			expectToken:   true, // Ожидаем не пустой токен
			expectedError: nil,
		},
		{
			name: "User exists during registration",
			dto:  model.AuthRequestDTO{Username: "existing_user", Password: "password"},
			mockSetup: func(mockRepo *mocks.RepoAuthInterface) {
				// Настраиваем мок для возврата пустого пользователя
				mockRepo.On("GetUser", mock.Anything, "existing_user", utils.GeneratePasswordHash("password")).
					Return(domain.User{}, nil)
				// Настраиваем мок для возврата ошибки ErrUserExist
				mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("domain.User")).
					Return(0, erorrs.ErrUserExist)
			},
			loggerSetup: func(mockLogger *mocks2.Logger) {
				// Настраиваем мок для вызова Error
				mockLogger.On("Error", mock.Anything, mock.Anything).Return()
			},
			expectToken:   false, // Токен не ожидается
			expectedError: fmt.Errorf("username is occupate: %w", erorrs.ErrUserExist),
		},
		{
			name: "Error during registration",
			dto:  model.AuthRequestDTO{Username: "invalid_user", Password: "password"},
			mockSetup: func(mockRepo *mocks.RepoAuthInterface) {
				// Настраиваем мок для возврата пустого пользователя
				mockRepo.On("GetUser", mock.Anything, "invalid_user", utils.GeneratePasswordHash("password")).
					Return(domain.User{}, nil)
				// Настраиваем мок для возврата ошибки
				mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("domain.User")).
					Return(0, errors.New("database error"))
			},
			loggerSetup: func(mockLogger *mocks2.Logger) {
				// Настраиваем мок для вызова Error с любыми аргументами
				mockLogger.On("Error", mock.Anything, mock.Anything).Return()
			},
			expectToken:   false, // Токен не ожидается
			expectedError: fmt.Errorf("registration failed: registration failed: %w", errors.New("database error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем моки
			tt.mockSetup(mockRepoAuth)
			if tt.loggerSetup != nil {
				tt.loggerSetup(mockLogger)
			}

			// Вызываем тестируемый метод
			token, err := authService.Authorization(context.Background(), tt.dto)

			// Проверяем результаты
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			// Проверяем, что токен не пустой, если ожидается
			if tt.expectToken {
				assert.NotEmpty(t, token, "Токен должен быть не пустым")
			} else {
				assert.Empty(t, token, "Токен должен быть пустым")
			}

			// Проверяем, что все ожидаемые вызовы были выполнены
			mockRepoAuth.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
