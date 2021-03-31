package main

import (
	"errors"
	"io"
	"log"

	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
	"github.com/pion/webrtc/v2/pkg/media/samplebuilder"
)

//MediaRoom is
type MediaRoom struct {
	RoomID          int
	API             *webrtc.API
	UserAudioTracks [MaxUserSize]*UserAudioTrack
	VideoTrack      *webrtc.Track
}

//NewMediaRoom is
func NewMediaRoom(id int) *MediaRoom {
	mediaEngine := NewMediaEngine()
	api := NewAPI(mediaEngine)

	userAudioTracks := CreateUserAudioTracks()
	videoTrack := NewVideoTrack()

	return &MediaRoom{
		RoomID:          id,
		API:             api,
		UserAudioTracks: userAudioTracks,
		VideoTrack:      videoTrack,
	}
}

//NewPeerConnection is
func (mediaRoom *MediaRoom) NewPeerConnection() *webrtc.PeerConnection {
	return NewPeerConnection(mediaRoom.API)
}

func (mediaRoom *MediaRoom) findSuitableAudioTrack(user *User) *UserAudioTrack {
	for _, userAudioTrack := range mediaRoom.UserAudioTracks {
		if userAudioTrack.User != nil {
			if userAudioTrack.User.ID == user.ID {
				log.Println("Used old audiotrack")
				userAudioTrack.User = user
				return userAudioTrack
			}
		} else {
			log.Println("Used new audio track")
			userAudioTrack.User = user
			return userAudioTrack
		}
	}
	return nil
}

func (mediaRoom *MediaRoom) findAudioTrackByUser(user *User) *UserAudioTrack {
	for _, userAudioTrack := range mediaRoom.UserAudioTracks {
		if userAudioTrack.User == nil {
			continue
		}
		if userAudioTrack.User.ID == user.ID {
			return userAudioTrack
		}
	}
	return nil
}

func (mediaRoom *MediaRoom) RemoveAudioTrackByUser(user *User) {
	for _, userAudioTrack := range mediaRoom.UserAudioTracks {
		if userAudioTrack.User == nil {
			continue
		}
		if userAudioTrack.User.ID == user.ID {
			userAudioTrack.User = nil
		}
	}
}

//JoinUser is
func (mediaRoom *MediaRoom) JoinUser(
	context *Context,
	sd webrtc.SessionDescription) (*webrtc.SessionDescription, error) {

	pc := context.RoomUser.PeerConnection

	userAudioTrack := mediaRoom.findSuitableAudioTrack(context.User)
	if userAudioTrack == nil {
		print("Room is full!")
		return nil, errors.New("Room is full")
	}

	mediaRoom.allowSendAudioTrackToAllPeers(context)

	if context.IsPublisher {
		log.Println("Publisher")
		AllowReceiveVideoTrack(pc)
		AllowReceiveAudioTrack(pc)

	} else {
		log.Println("Client")
		AllowSendVideoTrack(pc, mediaRoom.VideoTrack)
		AllowReceiveAudioTrack(pc)
	}

	mediaRoom.addOnTrack(context.RoomUser)

	log.Printf("Added RoomUser id : %d. RoomUser Count : %d", context.User.ID, len(context.Room.Users))

	DataChannelHandler(context)

	// Set the remote SessionDescription
	err := pc.SetRemoteDescription(sd)
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

func (mediaRoom *MediaRoom) allowSendAudioTrackToAllPeers(context *Context) {
	for _, userAudioTrack := range mediaRoom.UserAudioTracks {
		if userAudioTrack.IsSameUser(context.User) {
			AllowSendAudioTrack(context.RoomUser.PeerConnection, userAudioTrack.Track)
		}
	}
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
		userAudioTrack := mediaRoom.findAudioTrackByUser(roomUser.User)
		if userAudioTrack == nil {
			log.Println("User Audio track was nil!")
			break
		}

		audioTrack := userAudioTrack.Track
		if audioTrack == nil {
			log.Println("Audio track was nil!")
			break
		}
		if remoteTrack.Kind() == webrtc.RTPCodecTypeAudio {
			builderAudio.Push(rtp)
			for s := builderAudio.Pop(); s != nil; s = builderAudio.Pop() {
				mediaRoom.onAudioSample(roomUser, audioTrack, s)
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
	track *webrtc.Track,
	sample *media.Sample) {
	if err := track.WriteSample(*sample); err != nil && err != io.ErrClosedPipe {
		log.Panic(err)
	}
}

func (mediaRoom *MediaRoom) onVideoSample(
	roomUser *RoomUser,
	sample *media.Sample) {
	if err := mediaRoom.VideoTrack.WriteSample(*sample); err != nil && err != io.ErrClosedPipe {
		log.Panic(err)
	}
}

//AddStreamer is
func (mediaRoom *MediaRoom) AddStreamer() {

}

//AddViewer is
func (mediaRoom *MediaRoom) AddViewer() {

}
