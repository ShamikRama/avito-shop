package service

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/erorrs"
	mocks2 "avito-shop/internal/logger/mocks"
	"avito-shop/internal/model"
	"avito-shop/internal/service/mocks"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

//type FakeUser struct {
//	ID       int
//	Username string
//	Coins    int
//}
//
//func NewFakeUser() FakeUser {
//	return FakeUser{1, "user1", 10}
//}

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
		//fake          FakeUser
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
				mockRepo.On("BuyItem", mock.Anything, 1, domain.Item{ID: 1, Name: "cup", Price: 20}).Return(nil)
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
				mockRepo.On("GetItem", mock.Anything, "car").Return(domain.Item{}, erorrs.ErrNotFound)
			},
			mockLogger: func() {
				mockLogger.On("Error", mock.Anything, mock.Anything)
			},
			expectedError: erorrs.ErrNotFound,
		},
		{
			name:   "user not found",
			userID: 29,
			dto:    model.BuyItemRequestDTO{Item: "cup"},
			mockRepo: func() {
				mockRepo.On("GetItem", mock.Anything, "cup").Return(domain.Item{ID: 1, Name: "cup", Price: 20}, nil)
				mockRepo.On("BuyItem", mock.Anything, 29, domain.Item{ID: 1, Name: "cup", Price: 20}).Return(erorrs.ErrNotFound)
			},
			mockLogger: func() {
				mockLogger.On("Error", mock.Anything, mock.Anything)
			},
			expectedError: erorrs.ErrNotFound,
		},
		//{
		//	name: "balance is wrong",
		//	user: domain.User{ID: 1, Coins: 50},
		//	dto:  model.BuyItemRequestDTO{Item: "cup"},
		//	fake: FakeUser{ID: 1, Coins: 3},
		//	mockRepo: func() {
		//		mockRepo.On("GetItem", mock.Anything, "cup").Return(domain.Item{ID: 1, Name: "cup", Price: 20}, nil)
		//		mockRepo.On("BuyItem", mock.Anything, domain.User{ID: 1}, domain.Item{ID: 1, Name: "cup", Price: 20}).Return(erorrs.ErrInsufficientFunds)
		//	},
		//	mockLogger: func() {
		//		mockLogger.On("Error", mock.Anything, mock.Anything)
		//	},
		//	expectedError: erorrs.ErrInsufficientFunds,
		//},
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
