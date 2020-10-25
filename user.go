package main

import "errors"

//User is
type User struct {
	ID       int
	Name     string
	Password string
}

//UserRepository is
type UserRepository interface {
	LoginUser(*User) bool
	RegisterUser(*User) error
	GetUserByID(int) *User
}

var userRepository UserRepository = &UserRepositoryMock{lastUserID: -1}

//UserRepositoryMock is
type UserRepositoryMock struct {
	users      []*User
	lastUserID int
}

//RegisterUser is
func (repo *UserRepositoryMock) RegisterUser(user *User) error {
	for _, eachUser := range repo.users {
		if eachUser.Name == user.Name {
			return errors.New("user name already using")
		}
	}
	repo.lastUserID = repo.lastUserID + 1
	user.ID = repo.lastUserID
	repo.users = append(repo.users, user)
	return nil
}

//LoginUser is
func (repo *UserRepositoryMock) LoginUser(user *User) bool {
	for _, eachUser := range repo.users {
		if eachUser.Name == user.Name {
			if eachUser.Password == user.Password {
				user.ID = eachUser.ID
				return true
			}
		}
	}
	return false
}

//GetUserByID is
func (repo *UserRepositoryMock) GetUserByID(ID int) *User {
	for _, eachUser := range repo.users {
		if eachUser.ID == ID {
			return eachUser
		}
	}
	return nil
}
