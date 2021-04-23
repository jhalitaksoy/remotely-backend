package main

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

func CreateJWTToken(user *User) (string, error) {
	expiresAt := time.Now().Add(jwtExpireTime).Unix()
	claims := jwtClaims{
		UserID:   user.ID,
		Username: user.Name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Issuer:    "nameOfWebsiteHere", //TODO: change this
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}
