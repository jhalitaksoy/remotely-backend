package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pion/webrtc/v2"
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
	statusCode, userID := myContext.AuthService.Login(&loginParameters)
	c.String(statusCode, userID)
}

func createRoomRoute(c *gin.Context) {
	userID, err := strconv.Atoi(c.GetHeader("userID"))
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	user := myContext.UserStore.GeById(userID)

	room := Room{}
	err = c.BindJSON(&room)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	ok := roomRepository.CreateRoom(user, &room)

	if !ok {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	c.AbortWithStatus(http.StatusOK)
}

func joinRoomRoute(c *gin.Context) {
	userID, err := strconv.Atoi(c.GetHeader("userID"))
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	user := myContext.UserStore.GeById(userID)

	roomIDStr := c.Params.ByName("roomid")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}
	room := roomRepository.GetRoomByID(roomID)
	ok := roomRepository.JoinRoom(user, room)
	if !ok {
		c.AbortWithStatus(http.StatusConflict)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

func listRoomsRoute(c *gin.Context) {
	userID, err := strconv.Atoi(c.GetHeader("userID"))
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	user := myContext.UserStore.GeById(userID)

	rooms := roomRepository.ListRooms(user)
	if rooms == nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}
	c.JSON(http.StatusOK, rooms)
}

func chatRoomRoute(c *gin.Context) {
	//Security issue !!!!!!!!!!!!!!

	//userID, err := strconv.Atoi(c.GetHeader("userID"))
	//if err != nil {
	//	c.AbortWithStatus(http.StatusBadRequest)
	//	return
	//}
	//user := userRepository.GetUserByID(userID)

	roomIDStr := c.Params.ByName("roomid")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}
	room := roomRepository.GetRoomByID(roomID)
	c.JSON(http.StatusOK, room.ChatMessages)
}

func getRoomRoute(c *gin.Context) {
	//userID, err := strconv.Atoi(c.GetHeader("userID"))
	//if err != nil {
	//	c.AbortWithStatus(http.StatusBadRequest)
	//	return
	//}
	//user := userRepository.GetUserByID(userID)

	roomIDStr := c.Params.ByName("roomid")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}
	room := roomRepository.GetRoomByID(roomID)
	if room == nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}
	c.JSON(http.StatusOK, room)
}

type message struct {
	Name string                    `json:"name"`
	SD   webrtc.SessionDescription `json:"sd"`
}

func sdpRoute(c *gin.Context) {
	userID, err := strconv.Atoi(c.GetHeader("userID"))
	/*if err != nil {
		//c.AbortWithStatus(http.StatusBadRequest)
		//return
	}*/
	var user *User

	if err == nil {
		user = myContext.UserStore.GeById(userID)
		if user == nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	} else {
		_, err := myContext.UserStore.Create(&User{
			Name:      "anonymous_user_" + createUUID(),
			Anonymous: true,
		})
		if err != nil {
			log.Printf("Error when creating anonymous user : %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

	}

	roomIDStr := c.Params.ByName("roomid")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	room := roomRepository.GetRoomByID(roomID)
	if room == nil {
		c.AbortWithStatus(http.StatusBadRequest)
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

	answer, err := room.JoinUserToRoom(user, offer.SD, isPublisher)
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

func listUsers(c *gin.Context) {

	roomIDStr := c.Params.ByName("roomid")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}
	room := roomRepository.GetRoomByID(roomID)
	if room == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	userNames := []string{}
	for _, roomUser := range room.Users {
		userNames = append(userNames, roomUser.User.Name)
	}

	c.JSON(http.StatusOK, userNames)
}
