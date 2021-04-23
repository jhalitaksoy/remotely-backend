package main

import (
	"errors"
	"log"

	"github.com/go-pg/pg"
)

//UserStoreDatabaseImpl ...
type UserStoreDBImpl struct {
	DataBase *DataBase
}

func newUserStoreDBImpl(db *DataBase) *UserStoreDBImpl {
	return &UserStoreDBImpl{
		DataBase: db,
	}
}

func (userStore *UserStoreDBImpl) Create(user *User) (int, error) {
	userStored := userStore.GetByName(user.Name)
	if userStored != nil {
		return -1, errors.New("user name is not suitable")
	}

	_, err := userStore.DataBase.DB.Model(user).Insert()
	if err != nil {
		return -1, err
	}

	return user.ID, nil
}

func (userStore *UserStoreDBImpl) GeById(id int) *User {
	user := new(User)
	err := userStore.DataBase.DB.Model(user).Where("id = ?", id).Select()
	if err == pg.ErrNoRows {
		return nil
	}
	if err != nil {
		log.Printf("An error occured when selecting users %v", err)
		return nil
	}

	return user
}

func (userStore *UserStoreDBImpl) GetByName(name string) *User {
	user := new(User)
	err := userStore.DataBase.DB.Model(user).Where("name = ?", name).Select()
	if err == pg.ErrNoRows {
		return nil
	}
	if err != nil {
		log.Printf("An error occured when selecting users : %v", err)
		return nil
	}
	//todo : handle error (return)
	return user
}

func (userStore *UserStoreDBImpl) Update(user *User) {
	_, err := userStore.DataBase.DB.Model(user).Where("id = ?", user.ID).Update()

	if err != nil {
		log.Printf("An error occured when selecting users")
	}
	//todo : handle error (return)
}

func (userStore *UserStoreDBImpl) Delete(id int) {
	_, err := userStore.DataBase.DB.Model().Where("id = ?", id).Delete()

	if err != nil {
		log.Printf("An error occured when selecting users")
	}
	//todo : handle error (return)
}
