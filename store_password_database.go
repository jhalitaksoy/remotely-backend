package main

import (
	"errors"

	"github.com/go-pg/pg"
)

type Password struct {
	UserID int
	Hash   string
}

type PasswordStoreDBImpl struct {
	Database *DataBase
}

func newPasswordStoreDBImpl(DataBase *DataBase) *PasswordStoreDBImpl {
	return &PasswordStoreDBImpl{
		Database: DataBase,
	}
}

func (passwordStore *PasswordStoreDBImpl) Create(userID int, hash string) error {
	passwordStored := new(Password)
	err := passwordStore.Database.DB.Model(passwordStored).Where("user_id = ?", userID).Select()
	if err == pg.ErrNoRows {
		password := &Password{
			UserID: userID,
			Hash:   hash,
		}
		_, err = passwordStore.Database.DB.Model(password).Insert()

		if err != nil {
			return err
		}
		return nil
	} else {
		if err != nil {
			return err
		}
		return errors.New("there is a already password")
	}
}

func (passwordStore *PasswordStoreDBImpl) Get(userID int) *string {
	passwordStored := new(Password)
	err := passwordStore.Database.DB.Model(passwordStored).Where("user_id = ?", userID).Select()
	if err != nil {
		return nil
	}
	return &passwordStored.Hash
}

func (passwordStore *PasswordStoreDBImpl) Update(userID int, hash string) {
	password := &Password{
		UserID: userID,
		Hash:   hash,
	}
	_, _ = passwordStore.Database.DB.Model(password).Where("user_id = ?", userID).Update()
	/*if err != nil {
		//todo return error
		//return nil
	}*/
}

func (passwordStore *PasswordStoreDBImpl) Delete(userID int) {
	_, _ = passwordStore.Database.DB.Model().Where("user_id = ?", userID).Delete()
	/*if err != nil {
		//todo return error
		//return nil
	}*/
}
