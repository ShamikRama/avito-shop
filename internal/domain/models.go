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

// Transactions
type Transaction struct {
	ID         uint
	FromUserID uint
	ToUserID   uint
	Amount     int
}

// Item Catalog (для покупки предметов)
type ShopItem struct {
	ID    uint
	Name  string
	Price int
	Stock int
}
