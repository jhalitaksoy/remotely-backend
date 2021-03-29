package main

import (
	"flag"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const createTestUser = true

func main() {
	StartGin()
}

//StartGin is
func StartGin() *gin.Engine {
	r := gin.Default()

	// TODO : Look Before Production (Security)
	config := cors.DefaultConfig()
	config.AllowHeaders = append(config.AllowHeaders, "userid")
	config.AllowAllOrigins = true
	//config.AllowOrigins = []string{"http://localhost:3000"}
	r.Use(cors.New(config))

	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello :)")
	})

	r.GET("/ssl", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "https://remotely-sigma.vercel.app/")
		c.AbortWithStatus(http.StatusTemporaryRedirect)
	})
	user := r.Group("/user")
	{
		user.POST("/login", loginRoute)
		user.POST("/register", registerRoute)
	}

	r.Use(authRequired)
	{
		room := r.Group("/room")
		{
			room.POST("/get/:roomid", getRoomRoute)
			room.POST("/create", createRoomRoute)
			room.POST("/join/:roomid", joinRoomRoute)
			room.POST("/listRooms", listRoomsRoute)
			//Add Test
			room.POST("/chat/:roomid", chatRoomRoute)
		}

		stream := r.Group("/stream")
		{
			stream.POST("/sdp/:roomid", sdpRoute)
			//stream.POST("/audio/sdp/:roomid", sdpAudioRoute)
			//stream.POST("/publish/:roomid")
			//stream.POST("/join/:roomid")
		}
	}

	if createTestUser {
		println("Creating test users hlt and hlt2")

		userRepository.RegisterUser(&User{
			ID:       0,
			Name:     "hlt",
			Password: "asdfg",
		})

		userRepository.RegisterUser(&User{
			ID:       0,
			Name:     "hlt2",
			Password: "asdfg",
		})
	}

	if flag.Lookup("test.v") == nil {
		r.Run(":8080")
	} else {
		go r.Run(":8080")
		Info("Server Started")
	}
	return r
}
