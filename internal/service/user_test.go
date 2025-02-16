//go:build unit
// +build unit

package service

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/erorrs"
	mocks2 "avito-shop/internal/logger/mocks"
	"avito-shop/internal/model"
	"avito-shop/internal/service/mocks"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestUserService_SendCoinToUser(t *testing.T) {
	mockRepo := mocks.NewRepoUserInterface(t)
	mockLogger := mocks2.NewLogger(t)
	userService := NewUserService(mockRepo, mockLogger)

	tests := []struct {
		name          string
		fromUserID    int
		toUser        string
		amount        int
		mockRepo      func()
		mockLogger    func()
		expectedError error
	}{
		{
			name:       "success",
			fromUserID: 1,
			toUser:     "user2",
			amount:     500,
			mockRepo: func() {
				mockRepo.On("GetUserByName", mock.Anything, "user2").Return(2, nil)
				mockRepo.On("SendCoins", mock.Anything, 1, 2, 500).Return(nil)
			},
			mockLogger: func() {
				mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
			expectedError: nil,
		},
		{
			name:       "fromUser toUser ID is same",
			fromUserID: 1,
			toUser:     "user1",
			amount:     5,
			mockRepo: func() {
				mockRepo.On("GetUserByName", mock.Anything, "user1").Return(1, nil)
			},
			mockLogger: func() {
				mockLogger.On("Error", mock.Anything, mock.Anything)
			},
			expectedError: erorrs.ErrSelfTransfer,
		},
		{
			name:       "not enough money",
			fromUserID: 1,
			toUser:     "user2",
			amount:     2000,
			mockRepo: func() {
				mockRepo.On("GetUserByName", mock.Anything, "user2").Return(2, nil)
				mockRepo.On("SendCoins", mock.Anything, 1, 2, 2000).Return(erorrs.ErrInsufficientFunds)
			},
			mockLogger: func() {
				mockLogger.On("Error", mock.Anything, mock.Anything)
			},
			expectedError: erorrs.ErrInsufficientFunds,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockRepo()
			tt.mockLogger()

			err := userService.SendCoinToUser(context.Background(), tt.fromUserID, tt.toUser, tt.amount)

			assert.Equal(t, tt.expectedError, err)
			mockRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestUserService_BuyItem(t *testing.T) {
	mockRepo := mocks.NewRepoUserInterface(t)
	mockLogger := mocks2.NewLogger(t)
	userService := NewUserService(mockRepo, mockLogger)

	tests := []struct {
		name          string
		userID        int
		user          domain.User
		dto           model.BuyItemRequestDTO
		item          domain.Item
		mockRepo      func()
		mockLogger    func()
		expectedError error
	}{
		{
			name:   "success",
			userID: 1,
			dto:    model.BuyItemRequestDTO{Item: "cup"},
			mockRepo: func() {
				mockRepo.On("GetItem", mock.Anything, "cup").Return(domain.Item{ID: 1, Name: "cup", Price: 20}, nil)
				mockRepo.On("GetUser", mock.Anything, 1).Return(domain.User{ID: 1, Username: "user1", PasswordHash: "", Coins: 1000}, nil)
				mockRepo.On("BuyItem", mock.Anything, domain.User{ID: 1, Username: "user1", Coins: 1000}, domain.Item{ID: 1, Name: "cup", Price: 20}).Return(nil)
			},
			mockLogger: func() {
				mockLogger.On("Info", mock.Anything, mock.Anything)
			},
			expectedError: nil,
		},
		{
			name:   "item not found",
			userID: 1,
			dto:    model.BuyItemRequestDTO{Item: "car"},
			mockRepo: func() {
				mockRepo.On("GetItem", mock.Anything, "car").Return(domain.Item{ID: 45, Name: "cup", Price: 20324}, erorrs.ErrItemNotFound)
			},
			mockLogger: func() {
				mockLogger.On("Error", mock.Anything, mock.Anything)
			},
			expectedError: erorrs.ErrItemNotFound,
		},
		{
			name:   "user not found",
			userID: 25,
			dto:    model.BuyItemRequestDTO{Item: "cup"},
			mockRepo: func() {
				mockRepo.On("GetItem", mock.Anything, "cup").Return(domain.Item{ID: 1, Name: "cup", Price: 20}, nil)
				mockRepo.On("GetUser", mock.Anything, 25).Return(domain.User{ID: 0, Username: "", Coins: 0}, erorrs.ErrNotFound)
			},
			mockLogger: func() {
				mockLogger.On("Error", mock.Anything, mock.Anything)
			},
			expectedError: erorrs.ErrNotFound,
		},
		{
			name:   "balance is not enough",
			userID: 2,
			dto:    model.BuyItemRequestDTO{Item: "cup"},
			mockRepo: func() {
				mockRepo.On("GetItem", mock.Anything, "cup").Return(domain.Item{ID: 1, Name: "cup", Price: 20}, nil)
				mockRepo.On("GetUser", mock.Anything, 2).Return(domain.User{ID: 2, Username: "user2", Coins: 2}, nil)
			},
			mockLogger: func() {
				mockLogger.On("Error", mock.Anything, mock.Anything)
			},
			expectedError: erorrs.ErrInsufficientFunds,
		},
		{
			name:   "error not found",
			userID: 3,
			dto:    model.BuyItemRequestDTO{Item: "cup"},
			mockRepo: func() {
				mockRepo.On("GetItem", mock.Anything, "cup").Return(domain.Item{ID: 3, Name: "cup", Price: 20}, nil)
				mockRepo.On("GetUser", mock.Anything, 3).Return(domain.User{ID: 3, Username: "user3", Coins: 1000}, nil)
				mockRepo.On("BuyItem", mock.Anything, domain.User{ID: 3, Username: "user3", Coins: 1000}, domain.Item{ID: 1, Name: "cup", Price: 20}).Return(erorrs.ErrNotFound)
			},
			mockLogger: func() {
				mockLogger.On("Error", mock.Anything, mock.Anything)
			},
			expectedError: erorrs.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockRepo()
			tt.mockLogger()

			err := userService.BuyItem(context.Background(), tt.userID, tt.dto)

			assert.Equal(t, tt.expectedError, err)
			mockRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUserInfo(t *testing.T) {
	mockRepo := mocks.NewRepoUserInterface(t)
	mockLogger := mocks2.NewLogger(t)
	service := NewUserService(mockRepo, mockLogger)

	tests := []struct {
		name           string
		userID         int
		mockRepo       func(*mocks.RepoUserInterface)
		mockLogger     func(*mocks2.Logger)
		expectedResult model.InfoResponseDTO
		expectedError  error
	}{
		{
			name:   "successful get info",
			userID: 1,
			mockRepo: func(m *mocks.RepoUserInterface) {
				m.On("GetPurchasedItems", mock.Anything, 1).Return([]model.ItemDTO{
					{Type: "cup", Quantity: 1},
				},
					nil)
				m.On("GetUser", mock.Anything, 1).Return(domain.User{ID: 1, Username: "user1", Coins: 980},
					nil)
				m.On("GetCoinHistory", mock.Anything, 1, "user1").Return(model.CoinHistoryDTO{
					[]model.TransactionHistoryDTO{},
					[]model.TransactionHistoryDTO{},
				},
					nil)
			},
			mockLogger: func(l *mocks2.Logger) {
				l.On("Info", mock.Anything).Maybe()
			},
			expectedResult: model.InfoResponseDTO{
				Coins: 980,
				Inventory: []model.ItemDTO{
					{Type: "cup", Quantity: 1},
				},
				CoinHistory: model.CoinHistoryDTO{
					Received: []model.TransactionHistoryDTO{},
					Sent:     []model.TransactionHistoryDTO{},
				},
			},
			expectedError: nil,
		},
		{
			name:   "error no rows",
			userID: 2,
			mockRepo: func(m *mocks.RepoUserInterface) {
				m.On("GetPurchasedItems", mock.Anything, 2).Return(nil, erorrs.ErrItemNotFound)
				m.On("GetUser", mock.Anything, 2).Return(domain.User{}, nil)
				m.On("GetCoinHistory", mock.Anything, 2, "").Return(model.CoinHistoryDTO{}, nil)
			},
			mockLogger: func(l *mocks2.Logger) {
				l.On("Error", mock.Anything, mock.Anything).Maybe()
			},
			expectedResult: model.InfoResponseDTO{
				Inventory: []model.ItemDTO{},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockRepo(mockRepo)
			tt.mockLogger(mockLogger)

			result, err := service.GetUserInfo(context.Background(), tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
