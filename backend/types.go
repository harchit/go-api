package main

import (
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type TransferRequest struct {
	ToAccount int `json:"toAccount"`
	Amount    int `json:"amount"`
}

type UpdatePassword struct {
	NewPass string `json:"password"`
}

type Account struct {
	ID                int       `json:"id"`
	FirstName         string    `json:"firstName"`
	LastName          string    `json:"lastN ame"`
	Number            int64     `json:"number"`
	EncryptedPassword string    `json:"-"`
	Balance           int64     `json:"balance"`
	CreatedAt         time.Time `json:"createdAt"`
}

func (a *Account) ValidPassword(pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(a.EncryptedPassword), []byte(pw)) == nil
}

// its bad practice to use 'Account' as the actual request Account, because
// we're not providing the ID, balance. These are made in the backend
type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Password  string `json:"password"`
}

type LoginResponse struct {
	Number int64  `json:"number"`
	Token  string `json:"token"`
}

type LoginRequest struct {
	Number   int64  `json:"number"`
	Password string `json:"password"`
}

func NewAccount(firstName, lastName, password string) (*Account, error) {
	encpw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Account{
		FirstName:         firstName,
		LastName:          lastName,
		Number:            int64(rand.Intn(10000)),
		EncryptedPassword: string(encpw),
		CreatedAt:         time.Now().UTC(),
	}, nil
}
