package model

type AuthRequestDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SendCoinRequestDTO struct {
	ToUser string `json:"toUser" binding:"required"`
	Amount int    `json:"amount" binding:"required,gt=0"`
}

type ItemDTO struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type TransactionHistoryDTO struct {
	FromUser string `json:"fromUser,omitempty"`
	ToUser   string `json:"toUser,omitempty"`
	Amount   int    `json:"amount"`
}

type CoinHistoryDTO struct {
	Received []TransactionHistoryDTO `json:"received"`
	Sent     []TransactionHistoryDTO `json:"sent"`
}

type InfoResponseDTO struct {
	Coins       int            `json:"coins"`
	Inventory   []ItemDTO      `json:"inventory"`
	CoinHistory CoinHistoryDTO `json:"coinHistory"`
}

type ErrorResponseDTO struct {
	Error string `json:"error"`
}

type BuyItemRequestDTO struct {
	Item string `uri:"item" binding:"required"`
}
