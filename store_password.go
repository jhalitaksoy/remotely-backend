package main

import (
	"errors"
)

//PasswordStore ...
type PasswordStore interface {
	Create(userID int, hash string) error
	Get(userID int) *string
	Update(userID int, hash string)
	Delete(userID int)
}

type PasswordStoreImpl struct {
	Passwords map[int]*string
}

func newPasswordStoreImpl() *PasswordStoreImpl {
	return &PasswordStoreImpl{
		Passwords: make(map[int]*string),
	}
}

func (passwordStore *PasswordStoreImpl) Create(userID int, hash string) error {
	passwordStored := passwordStore.Passwords[userID]
	if passwordStored != nil {
		return errors.New("there is a already password")
	}

	passwordStore.Passwords[userID] = &hash

	return nil
}

func (passwordStore *PasswordStoreImpl) Get(userID int) *string {
	return passwordStore.Passwords[userID]
}

func (passwordStore *PasswordStoreImpl) Update(userID int, hash string) {
	passwordStore.Passwords[userID] = &hash
}

func (passwordStore *PasswordStoreImpl) Delete(userID int) {
	passwordStore.Passwords[userID] = nil
}
