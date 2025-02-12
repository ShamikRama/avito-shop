package erorrs

import (
	"errors"
)

var (
	ErrNotFound      = errors.New("user not found")
	ErrUserExist     = errors.New("user already exist")
	ErrSigningMethod = errors.New("invalid signing method")
)

var (
	ErrSelfTransfer      = errors.New("нельзя переводить самому себе")
	ErrInsufficientFunds = errors.New("недостаточно средств")
)
