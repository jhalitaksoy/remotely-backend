package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func authRequired(c *gin.Context) {
	userID, err := strconv.Atoi(c.GetHeader("userID"))
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		log.Print("User ID not found in header!")
		return
	}
	user := userRepository.GetUserByID(userID)
	if user == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		log.Print("User Not Found By ID!")
		return
	}

	c.Set("user", user)
}
