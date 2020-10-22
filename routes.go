package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

	c.JSON(http.StatusOK, rooms)
}
