package domain

type User struct {
	ID           int
	Username     string
	PasswordHash string
	Coins        int
}

type Item struct {
	ID    int
	Name  string
	Price int
}

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
}
