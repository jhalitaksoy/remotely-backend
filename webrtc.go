package main

import (
	"io"
	"log"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
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
	//rtpAverageFrameWidth = 7
)

func NewRTPOpusCodec() webrtc.RTPCodecParameters {
	return webrtc.RTPCodecParameters{
		RTPCodecCapability: NewAudioRTPCodecCapability(),
		PayloadType:        111,
	}
}

func NewAudioRTPCodecCapability() webrtc.RTPCodecCapability {
	return webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 48000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil}
}

func NewRTPVP8Codec() webrtc.RTPCodecParameters {
	return webrtc.RTPCodecParameters{
		RTPCodecCapability: NewVideoRTPCodecCapability(),
		PayloadType:        96,
	}
}

func NewVideoRTPCodecCapability() webrtc.RTPCodecCapability {
	return webrtc.RTPCodecCapability{MimeType: "video/VP8", ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil}
}

func NewMediaEngine() *webrtc.MediaEngine {
	mediaEngine := webrtc.MediaEngine{}
	audioCodec := NewRTPOpusCodec()
	videoCodec := NewRTPVP8Codec()
	err := mediaEngine.RegisterCodec(audioCodec, webrtc.RTPCodecTypeAudio)
	if err != nil {
		log.Println(err)
		return nil
	}
	err = mediaEngine.RegisterCodec(videoCodec, webrtc.RTPCodecTypeVideo)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &mediaEngine
}

func NewAPI(mediaEngine *webrtc.MediaEngine) *webrtc.API {
	return webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))
}

func NewAudioTrack() *webrtc.TrackLocalStaticRTP {
	audioTrack, err := webrtc.NewTrackLocalStaticRTP(NewAudioRTPCodecCapability(), "pion", "audio")
	if err != nil {
		panic(err)
	}
	return audioTrack
}

func NewVideoTrack() *webrtc.TrackLocalStaticRTP {
	videoTrack, err := webrtc.NewTrackLocalStaticRTP(NewVideoRTPCodecCapability(), "pion", "video")
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

func NewPeerConnection(api *webrtc.API) *webrtc.PeerConnection {
	pc, err := api.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		panic(err)
	}
	return pc
}

// Allow us to receive 1 audio track
func AllowReceiveAudioTrack(pc *webrtc.PeerConnection) {
	if _, err := pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}
}

// Allow us to receive 1 video track
func AllowReceiveVideoTrack(pc *webrtc.PeerConnection) {
	if _, err := pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}
}

//Allow us to send audio track to client
func AllowSendAudioTrack(pc *webrtc.PeerConnection, audioTrack *webrtc.TrackLocalStaticRTP) {
	_, err := pc.AddTrack(audioTrack)
	if err != nil {
		panic(err)
	}
}

//Allow us to send video track to client
func AllowSendVideoTrack(pc *webrtc.PeerConnection, videoTrack *webrtc.TrackLocalStaticRTP) {
	_, err := pc.AddTrack(videoTrack)
	if err != nil {
		panic(err)
	}
}

func sendPLIInterval(pc *webrtc.PeerConnection, remoteTrack *webrtc.TrackRemote) {
	// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
	// This can be less wasteful by processing incoming RTCP events, then we would emit a NACK/PLI when a viewer requests it
	go func() {
		ticker := time.NewTicker(rtcpPLIInterval)
		for range ticker.C {
			rtcpSendErr := pc.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())}})
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
	Track *webrtc.TrackLocalStaticRTP
}

func (userAudioTrack *UserAudioTrack) IsSameUser(user *User) bool {
	return userAudioTrack.User != nil && userAudioTrack.User.ID == user.ID
}
