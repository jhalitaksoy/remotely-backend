package main

import (
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

//AuthService ...
type AuthService interface {
	Register(registerParameters *RegisterParameters) int
	Login(loginParameters *LoginParameters) (int, string)
	SetUserStore(userStore UserStore)
	SetPasswordStore(passwordStore PasswordStore)
	GetUserStore() UserStore
	GetPasswordStore() PasswordStore
}

type AuthServiceImpl struct {
	userStore     UserStore
	passwordStore PasswordStore
}

func newAuthServiceImpl() *AuthServiceImpl {
	return &AuthServiceImpl{}
}

func (authService *AuthServiceImpl) SetUserStore(userStore UserStore) {
	authService.userStore = userStore
}
func (authService *AuthServiceImpl) SetPasswordStore(passwordStore PasswordStore) {
	authService.passwordStore = passwordStore
}

func (authService *AuthServiceImpl) GetUserStore() UserStore {
	return authService.userStore
}
func (authService *AuthServiceImpl) GetPasswordStore() PasswordStore {
	return authService.passwordStore
}

func (authService *AuthServiceImpl) Register(registerParameters *RegisterParameters) int {
	hash, err := hashAndSaltPassword(registerParameters.Password)
	if err != nil {
		log.Printf("An error occured while hashing password , error : %v\n", err)
		return http.StatusInternalServerError
	}

	_, err = authService.GetUserStore().Create(&User{
		Name: registerParameters.Name,
	})

	if err != nil {
		return http.StatusConflict
	}

	user := authService.GetUserStore().GetByName(registerParameters.Name)

	err = authService.GetPasswordStore().Create(user.ID, hash)
	if err != nil {
		log.Printf("An error occured when creating password : %v", err)
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func (authService *AuthServiceImpl) Login(loginParameters *LoginParameters) (int, string) {
	user := authService.GetUserStore().GetByName(loginParameters.Name)

	if user == nil {
		return http.StatusConflict, ""
	}

	hash := authService.GetPasswordStore().Get(user.ID)
	if hash == nil {
		return http.StatusConflict, ""
	}

	hashBytes := []byte(*hash)
	plainText := []byte(loginParameters.Password)

	err := bcrypt.CompareHashAndPassword(hashBytes, plainText)
	if err != nil {
		return http.StatusConflict, ""
	} else {
		jwt := authService.CreateJWTToken(user)
		return http.StatusOK, jwt
	}
}

func (authService *AuthServiceImpl) CreateJWTToken(user *User) string {
	signedToken, err := CreateJWTToken(user)
	if err != nil {
		panic(err)
	}
	return signedToken
}
