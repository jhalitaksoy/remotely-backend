package main

//User is
type User struct {
	ID   int
	Name string
}
//UserRepository is
type UserRepository interface {
	AddUser(*User) error
}

var userRepository UserRepository = &UserRepositoryMock{lastUserID: -1}

//UserRepositoryMock is
type UserRepositoryMock struct {
	users      []*User
	lastUserID int
}

//AddUser is
func (repo *UserRepositoryMock) AddUser(user *User) error {
	repo.lastUserID = repo.lastUserID + 1
	user.ID = repo.lastUserID
	repo.users = append(repo.users, user)
	return nil
}
