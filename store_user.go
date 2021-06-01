package main

import (
	"errors"
)

//UserStore ...
type UserStore interface {
	Create(user *User) (int, error)
	GeById(id int) *User
	GetByName(name string) *User
	Update(user *User)
	Delete(id int)
}

type UserStoreImpl struct {
	lastUserID int
	Users      map[int]*User
}

func newUserStoreImpl() *UserStoreImpl {
	return &UserStoreImpl{
		lastUserID: -1,
		Users:      make(map[int]*User),
	}
}

func (userStore *UserStoreImpl) Create(user *User) (int, error) {
	userStored := userStore.GetByName(user.Name)
	if userStored != nil {
		return -1, errors.New("user name is not suitable")
	}

	id := userStore.createNewUserId()

	user.ID = id

	userStore.Users[user.ID] = user
	return user.ID, nil
}

func (userStore *UserStoreImpl) GeById(id int) *User {
	return userStore.Users[id]
}

func (userStore *UserStoreImpl) GetByName(name string) *User {
	for _, user := range userStore.Users {
		if user.Name == name {
			return user
		}
	}

	return nil
}

func (userStore *UserStoreImpl) Update(user *User) {
	userStore.Users[user.ID] = user
}

func (userStore *UserStoreImpl) Delete(id int) {
	delete(userStore.Users, id)
}

func (userStore *UserStoreImpl) createNewUserId() int {
	userStore.lastUserID++
	return userStore.lastUserID
}
