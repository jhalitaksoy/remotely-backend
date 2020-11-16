package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/pion/webrtc/v2"
)

var peerConnectionConfig = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

const (
	rtcpPLIInterval = time.Second
	// mode for frames width per timestamp from a 30 second capture
	rtpAverageFrameWidth = 7
)

//MediaRoom is
type MediaRoom struct {
	RoomID int
	api    *webrtc.API
	track  *webrtc.Track
}

//NewMediaRoom is
func NewMediaRoom(id int) *MediaRoom {
	m := webrtc.MediaEngine{}
	codec := webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000)
	m.RegisterCodec(codec)
	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))
	track, err := webrtc.NewTrack(webrtc.DefaultPayloadTypeVP8, rand.Uint32(), "pion", "video", codec)
	if err != nil {
		panic(err)
	}
	return &MediaRoom{
		RoomID: id,
		api:    api,
		track:  track,
	}
}

//NewPeerConnection is
func (mediaRoom *MediaRoom) NewPeerConnection() *webrtc.PeerConnection {
	pc, err := mediaRoom.api.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		panic(err)
	}
	return pc
}

//AddStreamer is
func (mediaRoom *MediaRoom) AddStreamer() {

}

//AddViewer is
func (mediaRoom *MediaRoom) AddViewer() {

}

//MediaRoomRepository is
type MediaRoomRepository struct {
	mediaRooms map[int]*MediaRoom
}

//NewMediaRoomRepository is
func NewMediaRoomRepository() *MediaRoomRepository {
	return &MediaRoomRepository{
		mediaRooms: make(map[int]*MediaRoom),
	}
}

var mediaRoomRepository *MediaRoomRepository = NewMediaRoomRepository()

//GetMediaRoomByRoomID is
func (repo *MediaRoomRepository) GetMediaRoomByRoomID(id int) *MediaRoom {
	mediaRoom := repo.mediaRooms[id]
	if mediaRoom == nil {
		mediaRoom = NewMediaRoom(id)
		repo.mediaRooms[id] = mediaRoom
	}
	return mediaRoom
}

//MessageType is
type MessageType string

const (
	chatControl = "chatControl"
	chatMessage = "chatMessage"
)

//ChatMessageData is
type ChatMessageData struct {
	UserID   string
	UserName string
	Message  string
}

//DataChannelHandler is
func DataChannelHandler(pc *webrtc.PeerConnection, room *Room, roomUser *RoomUser) {
	// Register data channel creation handling
	pc.OnDataChannel(func(d *webrtc.DataChannel) {
		roomUser.DataChannel = d

		log.Printf("New DataChannel %s - %d\n", d.Label(), d.ID())

		d.OnOpen(func() {
			log.Printf("Open Data channel %s - %d\n", d.Label(), d.ID())
		})

		d.OnClose(func() {
			log.Printf("Closed Data channel %s - %d.\n", d.Label(), d.ID())
		})

		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			var msg1 ChatMessage
			err := json.Unmarshal(msg.Data, &msg1)
			if err != nil {
				log.Fatalf("An error occured when parsing coming message %v", err)
			}
			log.Println(msg1)

			SendChatMessage(room, &msg1)
		})
	})
}

//ChatMessage is
type ChatMessage struct {
	Text string `json:"text"`
	User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
}

//SendChatMessage is
func SendChatMessage(room *Room, chatMessage *ChatMessage) {
	user := userRepository.GetUserByID(chatMessage.User.ID)
	if user == nil {
		log.Fatalf("User not found! id : %d", chatMessage.User.ID)
		return
	}
	chatMessage.User.Name = user.Name

	room.addChatMessage(chatMessage)

	json, err := json.Marshal(chatMessage)
	if err != nil {
		log.Fatalf("An error occured at converting ChatMessage to json : %v", err)
		return
	}
	for _, roomUser := range room.Users {
		roomUser.DataChannel.Send(json)
	}
}
