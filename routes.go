package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pion/webrtc/v2"
)

func loginRoute(c *gin.Context) {
	user := User{}
	err := c.BindJSON(&user)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	ok := userRepository.LoginUser(&user)

	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.String(http.StatusOK, fmt.Sprintf("%d", user.ID))
}

func registerRoute(c *gin.Context) {
	user := User{}
	err := c.BindJSON(&user)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	err = userRepository.RegisterUser(&user)

	if err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	c.String(http.StatusOK, fmt.Sprintf("%d", user.ID))
}

func createRoomRoute(c *gin.Context) {
	userID, err := strconv.Atoi(c.GetHeader("userID"))
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	user := userRepository.GetUserByID(userID)

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
	user := userRepository.GetUserByID(userID)

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
	user := userRepository.GetUserByID(userID)

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
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	user := userRepository.GetUserByID(userID)
	if user == nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
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

	mediaRoom := mediaRoomRepository.GetMediaRoomByRoomID(room.ID)
	if mediaRoom == nil {
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

	answer, err := mediaRoom.AddUser(user, room, offer.SD, isPublisher)
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
