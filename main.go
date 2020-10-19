package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	StartGin()
}

//StartGin is
func StartGin()*gin.Engine{
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "Hello :)")
	})
	user := r.Group("/user")
	{
		user.POST("/login/:name", loginRoute)
	}
	go r.Run()
	return r
}
