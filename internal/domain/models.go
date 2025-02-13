package domain

// User
type User struct {
	ID           int
	Username     string
	PasswordHash string
	Coins        int
}

// Item
type Item struct {
	ID    int
	Name  string
	Price int
}

// Inventory
type InventoryItem struct {
	ID       uint
	UserID   uint
	ItemType string
	Quantity int
}

type TransferWithUsernames struct {
	FromUser string
	ToUser   string
	Amount   int
	DateTime string
}
