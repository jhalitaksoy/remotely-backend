package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg"
	"github.com/pion/webrtc/v3"
)

func handleRegister(c *gin.Context) {
	var registerParameters RegisterParameters
	err := c.BindJSON(&registerParameters)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	statusCode := myContext.AuthService.Register(&registerParameters)
	c.Status(statusCode)
}

func handleLogin(c *gin.Context) {
	var loginParameters LoginParameters
	err := c.BindJSON(&loginParameters)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	statusCode, loginResult := myContext.AuthService.Login(&loginParameters)
	c.JSON(statusCode, loginResult)
}

func createRoomRoute(c *gin.Context) {
	_jwtClaims, exists := c.Get(jwtClaimsKeyName)
	if !exists {
		log.Println("Jwt Claims not found")
		return
	}

	jwtClaims, ok := _jwtClaims.(*jwtClaims)
	if !ok {
		log.Println("Jwt Claims is not jwtClaims")
		return
	}

	room := RoomDB{}
	err := c.BindJSON(&room)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	room.OwnerID = jwtClaims.UserID

	_, err = myContext.RoomStore.Create(room)

	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	c.AbortWithStatus(http.StatusOK)
}

func joinRoomRoute(c *gin.Context) {
	userID := -1

	_jwtClaims, exists := c.Get(jwtClaimsKeyName)
	if exists {
		jwtClaims, ok := _jwtClaims.(*jwtClaims)
		if ok {
			userID = jwtClaims.UserID
		}
	}

	var user *User

	if userID >= 0 {
		user = myContext.UserStore.GeById(userID)
		if user == nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	} else {
		userID, err := myContext.UserStore.Create(&User{
			Name:      "anonymous_user_" + createUUID(),
			Anonymous: true,
		})
		if err != nil {
			log.Printf("Error when creating anonymous user : %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		user = myContext.UserStore.GeById(userID)

	}

	roomIDStr := c.Params.ByName("roomid")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	room, err := myContext.RoomProvider.GetFromCache(roomID)
	if err == pg.ErrNoRows {
		c.Status(http.StatusNotFound)
		return
	} else if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var joinParameters JoinParameters
	err = c.BindJSON(&joinParameters)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	err = room.JoinUserWithoutSDP(myContext, user, joinParameters.IsPublisher)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.AbortWithStatus(http.StatusOK)
}

func wsRoomRoute(c *gin.Context) {

	token := c.Query("token")
	if token == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	claims, err := ParseJWTKey(token)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	userID := claims.UserID

	roomIDStr := c.Params.ByName("roomid")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	room, err := myContext.RoomProvider.GetFromCache(roomID)
	if err == pg.ErrNoRows {
		c.Status(http.StatusNotFound)
		return
	} else if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	peer := room.getPeer(userID)

	upgrader.CheckOrigin = func(r *http.Request) bool {
		// TODO : Security !!!
		return true
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &UserClientWS{hub: myContext.Hub, conn: conn, send: make(chan []byte, 256), Peer: peer}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

func listRoomsRoute(c *gin.Context) {

	_jwtClaims, exists := c.Get(jwtClaimsKeyName)
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		log.Println("Jwt Claims not found")
		return
	}

	jwtClaims, ok := _jwtClaims.(*jwtClaims)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		log.Println("Jwt Claims is not jwtClaims")
		return
	}

	rooms, err := myContext.RoomStore.GetByUserID(jwtClaims.UserID)
	if err == pg.ErrNoRows {
		rooms = make([]*RoomDB, 0)
	} else if err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	c.JSON(http.StatusOK, rooms)
}

func chatRoomRoute(c *gin.Context) {
	//Security issue !!!!!!!!!!!!!!

	/*_jwtClaims, exists := c.Get(jwtClaimsKeyName)
	if !exists {
		log.Println("Jwt Claims not found")
		return
	}

	jwtClaims, ok := _jwtClaims.(jwtClaims)
	if !ok {
		log.Println("Jwt Claims is not jwtClaims")
		return
	}*/

	roomIDStr := c.Params.ByName("roomid")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	room, err := myContext.RoomProvider.GetFromCache(roomID)
	if err == pg.ErrNoRows {
		c.Status(http.StatusNotFound)
		return
	} else if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, room.ChatMessages)
}

func getRoomRoute(c *gin.Context) {

	/*_jwtClaims, exists := c.Get(jwtClaimsKeyName)
	if !exists {
		log.Println("Jwt Claims not found")
		return
	}

	jwtClaims, ok := _jwtClaims.(jwtClaims)
	if !ok {
		log.Println("Jwt Claims is not jwtClaims")
		return
	}*/

	roomIDStr := c.Params.ByName("roomid")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}
	room, err := myContext.RoomStore.GetByID(roomID)
	if err == pg.ErrNoRows {
		c.Status(http.StatusNotFound)
		return
	} else if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, room)
}

type message struct {
	Name string                    `json:"name"`
	SD   webrtc.SessionDescription `json:"sd"`
}

func sdpRoute(c *gin.Context) {
	userID := -1

	_jwtClaims, exists := c.Get(jwtClaimsKeyName)
	if exists {
		jwtClaims, ok := _jwtClaims.(*jwtClaims)
		if ok {
			userID = jwtClaims.UserID
		}
	}

	var user *User

	if userID >= 0 {
		user = myContext.UserStore.GeById(userID)
		if user == nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	} else {
		userID, err := myContext.UserStore.Create(&User{
			Name:      "anonymous_user_" + createUUID(),
			Anonymous: true,
		})
		if err != nil {
			log.Printf("Error when creating anonymous user : %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		user = myContext.UserStore.GeById(userID)

	}

	roomIDStr := c.Params.ByName("roomid")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	room, err := myContext.RoomProvider.GetFromCache(roomID)
	if err == pg.ErrNoRows {
		c.Status(http.StatusNotFound)
		return
	} else if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var offer message
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&offer); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var isPublisher bool

	if offer.Name == "Publisher" {
		isPublisher = true
	} else if strings.HasPrefix(offer.Name, "Client") {
		isPublisher = false
	} else {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	answer, err := room.JoinUserToRoom(myContext, user, offer.SD, isPublisher)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusAccepted, map[string]interface{}{
		"Result": "Successfully received incoming client SDP",
		"SD":     answer,
	})
}

// Only Test Remove Later

func listUsers(c *gin.Context) {

	roomIDStr := c.Params.ByName("roomid")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}
	//calling GetFromCache is problem !!!
	room, err := myContext.RoomProvider.GetFromCache(roomID)
	if err == pg.ErrNoRows {
		c.Status(http.StatusNotFound)
		return
	} else if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	userNames := []string{}
	for _, roomUser := range room.Users {
		userNames = append(userNames, roomUser.User.Name)
	}

	c.JSON(http.StatusOK, userNames)
}

func listRoomCache(c *gin.Context) {
	//only test
	userNames := []string{}
	for _, room := range myContext.RoomProvider.(*RoomProviderImpl).Rooms {
		userNames = append(userNames, room.Name)
	}

	c.JSON(http.StatusOK, userNames)
}

// Only Test Remove Later
