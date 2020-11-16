package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pion/rtcp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media/samplebuilder"
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

	pc := mediaRoom.NewPeerConnection()

	switch strings.Split(offer.Name, ":")[0] {
	case "Publisher":
		// Allow us to receive 1 video track
		if _, err = pc.AddTransceiver(webrtc.RTPCodecTypeVideo); err != nil {
			panic(err)
		}

		// Set a handler for when a new remote track starts
		// Add the incoming track to the list of tracks maintained in the server
		addOnTrack(pc, mediaRoom.track)

		log.Println("Publisher")
	case "Client":
		_, err = pc.AddTrack(mediaRoom.track)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		log.Println("Client")
	default:
		c.AbortWithStatus(http.StatusBadRequest)
		//handler.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	roomUser := &RoomUser{
		User:           user,
		PeerConnection: pc,
	}

	room.addRoomUser(roomUser)

	log.Printf("Added RoomUser id : %d. RoomUser Count : %d", userID, len(room.Users))

	DataChannelHandler(pc, room, roomUser)

	// Set the remote SessionDescription
	err = pc.SetRemoteDescription(offer.SD)
	if err != nil {
		panic(err)
	}

	// Create answer
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = pc.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusAccepted, map[string]interface{}{
		"Result": "Successfully received incoming client SDP",
		"SD":     answer,
	})

	//c.JSON(http.StatusOK, rooms)
}

func addOnTrack(pc *webrtc.PeerConnection, localTrack *webrtc.Track) {
	// Set a handler for when a new remote track starts, this just distributes all our packets
	// to connected peers
	pc.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
		// This can be less wasteful by processing incoming RTCP events, then we would emit a NACK/PLI when a viewer requests it
		go func() {
			ticker := time.NewTicker(rtcpPLIInterval)
			for range ticker.C {
				rtcpSendErr := pc.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: remoteTrack.SSRC()}})
				if rtcpSendErr != nil {
					if rtcpSendErr == io.ErrClosedPipe {
						return
					}
					log.Println(rtcpSendErr)
				}
			}
		}()

		log.Println("Track acquired", remoteTrack.Kind(), remoteTrack.Codec())

		builder := samplebuilder.New(rtpAverageFrameWidth*5, &codecs.VP8Packet{})
		for {
			rtp, err := remoteTrack.ReadRTP()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Panic(err)
			}

			builder.Push(rtp)
			for s := builder.Pop(); s != nil; s = builder.Pop() {
				if err := localTrack.WriteSample(*s); err != nil && err != io.ErrClosedPipe {
					log.Panic(err)
				}
			}
		}
	})

}
