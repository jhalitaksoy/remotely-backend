package main

import (
	"io"
	"log"
	"math/rand"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v2"
)

const (
	//MaxUserSize is
	MaxUserSize = 10
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

func NewRTPOpusCodec() *webrtc.RTPCodec {
	return webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000)
}

func NewRTPVP8Codec() *webrtc.RTPCodec {
	return webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000)
}

func NewMediaEngine() *webrtc.MediaEngine {
	mediaEngine := webrtc.MediaEngine{}
	audioCodec := NewRTPOpusCodec()
	videoCodec := NewRTPVP8Codec()
	mediaEngine.RegisterCodec(audioCodec)
	mediaEngine.RegisterCodec(videoCodec)
	return &mediaEngine
}

func NewAPI(mediaEngine *webrtc.MediaEngine) *webrtc.API {
	return webrtc.NewAPI(webrtc.WithMediaEngine(*mediaEngine))
}

func NewAudioTrack() *webrtc.Track {
	audioCodec := NewRTPOpusCodec()
	audioTrack, err := webrtc.NewTrack(webrtc.DefaultPayloadTypeOpus, rand.Uint32(), "pion", "audio", audioCodec)
	if err != nil {
		panic(err)
	}
	return audioTrack
}

func NewVideoTrack() *webrtc.Track {
	videoCodec := NewRTPOpusCodec()
	videoTrack, err := webrtc.NewTrack(webrtc.DefaultPayloadTypeVP8, rand.Uint32(), "pion", "video", videoCodec)
	if err != nil {
		panic(err)
	}
	return videoTrack
}

func CreateUserAudioTracks() [MaxUserSize]*UserAudioTrack {
	userAudioTracks := [MaxUserSize]*UserAudioTrack{}

	for i := 0; i < MaxUserSize; i++ {
		audioTrack := NewAudioTrack()
		userAudioTrack := &UserAudioTrack{Track: audioTrack}
		userAudioTracks[i] = userAudioTrack
	}

	return userAudioTracks
}

func OnPeerConnectionChange(room *Room, mediaRoom *MediaRoom, roomUser *RoomUser, pcs webrtc.PeerConnectionState) {
	switch pcs {
	case webrtc.PeerConnectionStateNew:
	case webrtc.PeerConnectionStateConnecting:
	case webrtc.PeerConnectionStateConnected:
	case webrtc.PeerConnectionStateFailed:
		ClearUser(room, mediaRoom, roomUser)
	case webrtc.PeerConnectionStateClosed:
		ClearUser(room, mediaRoom, roomUser)
	case webrtc.PeerConnectionStateDisconnected:
		ClearUser(room, mediaRoom, roomUser)
	}
}

func ClearUser(room *Room, mediaRoom *MediaRoom, roomUser *RoomUser) {
	log.Println("ClearUser()")
	mediaRoom.RemoveAudioTrackByUser(roomUser.User)
	room.RemoveRoomUser(roomUser)
	if roomUser.User.Anonymous {
		userRepository.DeleteUser(roomUser.User.ID)
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

// UserAudioTrack is
type UserAudioTrack struct {
	User  *User
	Track *webrtc.Track
}
