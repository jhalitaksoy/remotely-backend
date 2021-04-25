package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var secretKey string

const jwtExpireTime = time.Hour * 24

var myContext *MyContext

func main() {

	LoadEnviromentVariables()

	myContext = newMyContext()

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = ":8080"
	}

	startGin(port)
}

func LoadEnviromentVariables() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error : %v\n", err)
	}

	secretKey = os.Getenv("SECRET_KEY")
	if len(secretKey) == 0 {
		panic("Secret key emty")
	}
}

func CorsConfig() cors.Config {
	return cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
		AllowAllOrigins:  true, //Change later
	}
}

func startGin(port string) *gin.Engine {
	r := gin.Default()

	// TODO : Look Before Production (Security)
	/*config := cors.DefaultConfig()
	config.AllowHeaders = append(config.AllowHeaders, "userid")
	config.AllowAllOrigins = true
	//config.AllowOrigins = []string{"http://localhost:3000"}*/
	r.Use(cors.New(CorsConfig()))

	r.GET("/", func(c *gin.Context) {
		c.String(200, "Server is Running!")
	})

	user := r.Group("/user")
	{
		user.POST("/login", handleLogin)
		user.POST("/register", handleRegister)
	}

	stream := r.Group("/stream")
	{
		stream.POST("/sdp/:roomid", sdpRoute)
	}

	//TODO : Remove later
	test := r.Group("/test")
	{
		test.GET("/users/:roomid", listUsers)
		test.GET("/listRoomCache", listRoomCache)
	}

	room := r.Group("/room")
	{
		room.POST("/get/:roomid", getRoomRoute)
		//Add Test
		room.POST("/chat/:roomid", chatRoomRoute)
	}

	r.Use(requiredAuthentication)
	{
		room := r.Group("/room")
		{
			room.POST("/create", createRoomRoute)
			//room.POST("/join/:roomid", joinRoomRoute)
			room.POST("/listRooms", listRoomsRoute)
		}

		stream := r.Group("/stream_private")
		{
			stream.POST("/sdp/:roomid", sdpRoute)
		}
	}

	/*if createTestUser {
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
	}*/

	if flag.Lookup("test.v") == nil {
		r.Run(port)
	} else {
		go r.Run(port)
		Info("Server Started")
	}
	return r
}
