package main

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
	"github.com/pion/webrtc/v2/pkg/media/samplebuilder"
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
	RoomID      int
	api         *webrtc.API
	audioTrack1 *webrtc.Track
	audioTrack2 *webrtc.Track
	videoTrack  *webrtc.Track
	audioCodec  *webrtc.RTPCodec
	videoCodec  *webrtc.RTPCodec
}

//NewMediaRoom is
func NewMediaRoom(id int) *MediaRoom {
	m := webrtc.MediaEngine{}

	codecVideo := webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000)
	m.RegisterCodec(codecVideo)
	codecAudio := webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000)
	m.RegisterCodec(codecAudio)

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))

	audioTrack1, err := webrtc.NewTrack(webrtc.DefaultPayloadTypeOpus, rand.Uint32(), "pion", "audio", codecAudio)
	if err != nil {
		panic(err)
	}

	audioTrack2, err := webrtc.NewTrack(webrtc.DefaultPayloadTypeOpus, rand.Uint32(), "pion", "audio", codecAudio)
	if err != nil {
		panic(err)
	}

	videoTrack, err := webrtc.NewTrack(webrtc.DefaultPayloadTypeVP8, rand.Uint32(), "pion", "video", codecVideo)
	if err != nil {
		panic(err)
	}

	return &MediaRoom{
		RoomID:      id,
		api:         api,
		audioTrack1: audioTrack1,
		audioTrack2: audioTrack2,
		videoTrack:  videoTrack,
		audioCodec:  codecAudio,
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

func (mediaRoom *MediaRoom) newAudioTrack() (*webrtc.Track, error) {
	return webrtc.NewTrack(webrtc.DefaultPayloadTypeOpus, rand.Uint32(), "pion", "audio", mediaRoom.audioCodec)
}

//AddUser is
func (mediaRoom *MediaRoom) AddUser(
	user *User,
	room *Room,
	sd webrtc.SessionDescription,
	isPublisher bool) (*webrtc.SessionDescription, error) {

	pc := mediaRoom.NewPeerConnection()

	audioTrack, err := mediaRoom.newAudioTrack()
	if err != nil {
		log.Panic(err)
	}

	roomUser := &RoomUser{
		User:           user,
		PeerConnection: pc,
		AudioTrack:     audioTrack,
	}

	room.addRoomUser(roomUser)
	log.Printf("Added RoomUser id : %d. RoomUser Count : %d", user.ID, len(room.Users))
	DataChannelHandler(pc, room, roomUser)

	//if roomUser.User.Name == "hlt" {
	//	_, err = pc.AddTrack(mediaRoom.audioTrack2)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//} else {
	//	_, err = pc.AddTrack(mediaRoom.audioTrack1)
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	for _, eachRoomUser := range room.Users {
		if eachRoomUser.User.ID != user.ID {
			_, err := pc.AddTrack(eachRoomUser.AudioTrack)
			if err != nil {
				return nil, err
			}
			eachRoomUser.PeerConnection.AddTrack(roomUser.AudioTrack)
		}
	}

	if isPublisher {
		log.Println("Publisher")

		// Allow us to receive 1 video track
		if _, err := pc.AddTransceiver(webrtc.RTPCodecTypeVideo); err != nil {
			panic(err)
		}

		// Allow us to receive 1 audio track
		if _, err := pc.AddTransceiver(webrtc.RTPCodecTypeAudio); err != nil {
			panic(err)
		}

	} else {
		log.Println("Client")

		_, err := pc.AddTrack(mediaRoom.videoTrack)
		if err != nil {
			return nil, err
		}

		// Allow us to receive 1 audio track
		if _, err := pc.AddTransceiver(webrtc.RTPCodecTypeAudio); err != nil {
			panic(err)
		}

	}
	mediaRoom.addOnTrack(roomUser)
	//if roomUser.User.Name == "hlt" {
	//	mediaRoom.addOnTrack(roomUser)
	//
	//} else {
	//	mediaRoom.addOnTrack2(roomUser)
	//}

	// Set the remote SessionDescription
	err = pc.SetRemoteDescription(sd)
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

	return &answer, nil
}

func (mediaRoom *MediaRoom) addOnTrack(roomUser *RoomUser) {
	// Set a handler for when a new remote track starts, this just distributes all our packets
	// to connected peers
	roomUser.PeerConnection.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		log.Println("Track acquired", remoteTrack.Kind(), remoteTrack.Codec())

		mediaRoom.onTrack(roomUser, remoteTrack, receiver)
	})
}

func (mediaRoom *MediaRoom) onTrack(roomUser *RoomUser, remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
	// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
	// This can be less wasteful by processing incoming RTCP events, then we would emit a NACK/PLI when a viewer requests it
	sendPLIInterval(roomUser.PeerConnection, remoteTrack)

	builderVideo := samplebuilder.New(rtpAverageFrameWidth*5, &codecs.VP8Packet{})
	builderAudio := samplebuilder.New(rtpAverageFrameWidth*5, &codecs.OpusPacket{})

	for {
		rtp, err := remoteTrack.ReadRTP()
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Panic(err)
		}

		if remoteTrack.Kind() == webrtc.RTPCodecTypeAudio {
			builderAudio.Push(rtp)
			for s := builderAudio.Pop(); s != nil; s = builderAudio.Pop() {
				mediaRoom.onAudioSample(roomUser, s)
			}
		} else if remoteTrack.Kind() == webrtc.RTPCodecTypeVideo {
			builderVideo.Push(rtp)
			for s := builderVideo.Pop(); s != nil; s = builderVideo.Pop() {
				mediaRoom.onVideoSample(roomUser, s)
			}
		}
	}
}

func (mediaRoom *MediaRoom) onAudioSample(
	roomUser *RoomUser,
	sample *media.Sample) {
	if err := roomUser.AudioTrack.WriteSample(*sample); err != nil && err != io.ErrClosedPipe {
		log.Panic(err)
	}
}

func (mediaRoom *MediaRoom) onVideoSample(
	roomUser *RoomUser,
	sample *media.Sample) {
	if err := mediaRoom.videoTrack.WriteSample(*sample); err != nil && err != io.ErrClosedPipe {
		log.Panic(err)
	}
}

func (mediaRoom *MediaRoom) addOnTrack2(roomUser *RoomUser) {
	// Set a handler for when a new remote track starts, this just distributes all our packets
	// to connected peers
	roomUser.PeerConnection.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		log.Println("Track acquired", remoteTrack.Kind(), remoteTrack.Codec())

		mediaRoom.onTrack2(roomUser, remoteTrack, receiver)
	})
}

func (mediaRoom *MediaRoom) onTrack2(roomUser *RoomUser, remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
	// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
	// This can be less wasteful by processing incoming RTCP events, then we would emit a NACK/PLI when a viewer requests it
	sendPLIInterval(roomUser.PeerConnection, remoteTrack)

	builderVideo := samplebuilder.New(rtpAverageFrameWidth*5, &codecs.VP8Packet{})
	builderAudio := samplebuilder.New(rtpAverageFrameWidth*5, &codecs.OpusPacket{})

	for {
		rtp, err := remoteTrack.ReadRTP()
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Panic(err)
		}

		if remoteTrack.Kind() == webrtc.RTPCodecTypeAudio {
			builderAudio.Push(rtp)
			for s := builderAudio.Pop(); s != nil; s = builderAudio.Pop() {
				mediaRoom.onAudioSample2(roomUser, s)
			}
		} else if remoteTrack.Kind() == webrtc.RTPCodecTypeVideo {
			builderVideo.Push(rtp)
			for s := builderVideo.Pop(); s != nil; s = builderVideo.Pop() {
				mediaRoom.onVideoSample2(roomUser, s)
			}
		}
	}
}

func (mediaRoom *MediaRoom) onAudioSample2(
	roomUser *RoomUser,
	sample *media.Sample) {
	if err := mediaRoom.audioTrack2.WriteSample(*sample); err != nil && err != io.ErrClosedPipe {
		log.Panic(err)
	}
}

func (mediaRoom *MediaRoom) onVideoSample2(
	roomUser *RoomUser,
	sample *media.Sample) {
	if err := mediaRoom.videoTrack.WriteSample(*sample); err != nil && err != io.ErrClosedPipe {
		log.Panic(err)
	}
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

func sendPLIInterval(pc *webrtc.PeerConnection, remoteTrack *webrtc.Track) {
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

}
