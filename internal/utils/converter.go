package utils

import (
	"avito-shop/internal/domain"
	"avito-shop/internal/model"
	"time"
)

// AuthRequestDTO -> domain.User (для регистрации)
func AuthRequestToUser(dto model.AuthRequestDTO) domain.User {
	return domain.User{
		Username:     dto.Username,
		PasswordHash: dto.Password,
	}
}

// domain.User -> AuthResponseDTO
func UserToAuthResponse(user domain.User, token string) model.AuthResponseDTO {
	return model.AuthResponseDTO{
		Token: token,
	}
}

// SendCoinRequestDTO -> domain.Transaction
func SendCoinRequestToTransaction(dto model.SendCoinRequestDTO, fromUserID uint) domain.Transaction {
	return domain.Transaction{
		FromUserID: fromUserID,
		Amount:     dto.Amount,
		CreatedAt:  time.Now(),
	}
}

// domain.InventoryItem -> ItemDTO
func InventoryItemToDTO(item domain.InventoryItem) model.ItemDTO {
	return model.ItemDTO{
		Type:     item.ItemType,
		Quantity: item.Quantity,
	}
}

// domain.Transaction -> TransactionHistoryDTO
func TransactionToHistoryDTO(t domain.Transaction, fromUsername, toUsername string) model.TransactionHistoryDTO {
	return model.TransactionHistoryDTO{
		FromUser: fromUsername,
		ToUser:   toUsername,
		Amount:   t.Amount,
		DateTime: t.CreatedAt.Format(time.RFC3339),
	}
}

// domain.User и связанные данные -> InfoResponseDTO
func UserToInfoResponse(
	user domain.User,
	inventory []domain.InventoryItem,
	received []domain.Transaction,
	sent []domain.Transaction,
	getUserName func(uint) string,
) model.InfoResponseDTO {
	response := model.InfoResponseDTO{
		Coins:     user.Coins,
		Inventory: make([]model.ItemDTO, 0, len(inventory)),
		CoinHistory: model.CoinHistoryDTO{
			Received: make([]model.TransactionHistoryDTO, 0, len(received)),
			Sent:     make([]model.TransactionHistoryDTO, 0, len(sent)),
		},
	}

	// Конвертация инвентаря
	for _, item := range inventory {
		response.Inventory = append(response.Inventory, InventoryItemToDTO(item))
	}

	// Конвертация истории транзакций
	for _, t := range received {
		response.CoinHistory.Received = append(response.CoinHistory.Received,
			TransactionToHistoryDTO(t, getUserName(t.FromUserID), getUserName(t.ToUserID)))
	}

	for _, t := range sent {
		response.CoinHistory.Sent = append(response.CoinHistory.Sent,
			TransactionToHistoryDTO(t, getUserName(t.FromUserID), getUserName(t.ToUserID)))
	}

	return response
}

// BuyItemRequestDTO -> domain.ShopItem (частичное преобразование)
func BuyRequestToShopItem(dto model.BuyItemRequestDTO) domain.ShopItem {
	return domain.ShopItem{
		Name: dto.Item,
	}
}
