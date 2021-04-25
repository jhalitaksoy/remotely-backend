package main

import (
	"log"
	"net/http"

	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

const jwtClaimsKeyName = "jwtClaims"

func requiredAuthentication(context *gin.Context) {
	header := context.Request.Header.Get("Authorization")
	if len(header) == 0 {
		log.Println("Authorization value is emty")
		context.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	array := strings.Split(header, " ")
	if len(array) != 2 {
		log.Println("Authorization value is not suitable")
		context.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	jwtFromHeader := array[1]
	token, err := jwt.ParseWithClaims(
		jwtFromHeader,
		&jwtClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		},
	)
	if err != nil {
		log.Println("Can not parse jwt token")
		context.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		log.Println("couldn't parse claims")
		return
	}

	context.Set(jwtClaimsKeyName, claims)

	user := myContext.UserStore.GeById(claims.UserID)
	if user == nil {
		context.AbortWithStatus(http.StatusUnauthorized)
		log.Print("User Not Found By ID!")
		return
	}

	context.Set("user", user)

	context.Next()
}

type jwtClaims struct {
	UserID   int    `json:"userid"`
	Username string `json:"username"`
	jwt.StandardClaims
}
