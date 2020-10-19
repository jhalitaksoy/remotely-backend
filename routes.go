package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func loginRoute(c *gin.Context) {
	userName := c.Params.ByName("name")
	user := User{Name: userName}
	err := userRepository.AddUser(&user)
	if  err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}
	c.String(http.StatusOK, fmt.Sprintf("%d", user.ID))
}
