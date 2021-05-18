package main

import (
	"errors"
	"log"
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

func ParseJWTKey(key string) (*jwtClaims, error) {
	token, err := jwt.ParseWithClaims(
		key,
		&jwtClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		},
	)
	if err != nil {
		log.Println("Can not parse jwt token")
		return nil, err
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		log.Println("couldn't parse claims")
		return nil, errors.New("couldn't parse claims")
	}
	return claims, nil
}
