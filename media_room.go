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
	RoomID int
	api    *webrtc.API
	//audioTrack *webrtc.Track
	UserAudioTracks [MaxUserSize]*UserAudioTrack
	videoTrack      *webrtc.Track
}

//NewMediaRoom is
func NewMediaRoom(id int) *MediaRoom {
	mediaEngine := NewMediaEngine()
	api := NewAPI(mediaEngine)

	userAudioTracks := CreateUserAudioTracks()
	videoTrack := NewVideoTrack()

	return &MediaRoom{
		RoomID: id,
		api:    api,
		//audioTrack: audioTrack,
		UserAudioTracks: userAudioTracks,
		videoTrack:      videoTrack,
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

func (mediaRoom *MediaRoom) findSuitableAudioTrack(userID int) *UserAudioTrack {
	for _, userAudioTrack := range mediaRoom.UserAudioTracks {
		if userAudioTrack.User != nil {
			if userAudioTrack.User.ID == userID {
				log.Println("Used old track")
				return userAudioTrack
			}
		} else {
			log.Println("Used new track")
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

//AddUser is
func (mediaRoom *MediaRoom) AddUser(
	user *User,
	room *Room,
	sd webrtc.SessionDescription,
	isPublisher bool) (*webrtc.SessionDescription, error) {
	pc := mediaRoom.NewPeerConnection()

	roomUser := NewRoomUser(user, pc)

	room.addRoomUser(roomUser)

	pc.OnConnectionStateChange(func(pcs webrtc.PeerConnectionState) {
		log.Println(pcs.String())
		OnPeerConnectionChange(room, mediaRoom, roomUser, pcs)
	})

	userAudioTrack := mediaRoom.findSuitableAudioTrack(user.ID)
	if userAudioTrack == nil {
		print("Room is full!")
		return nil, errors.New("Room is full")
	}

	userAudioTrack.User = user

	for _, userAudioTrack := range mediaRoom.UserAudioTracks {
		if userAudioTrack.User == nil || userAudioTrack.User.ID != user.ID {
			_, err := pc.AddTrack(userAudioTrack.Track)
			if err != nil {
				panic(err)
			}
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

		mediaRoom.addOnTrack(roomUser)

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

		mediaRoom.addOnTrack(roomUser)
	}

	log.Printf("Added RoomUser id : %d. RoomUser Count : %d", user.ID, len(room.Users))

	DataChannelHandler(pc, room, mediaRoom, roomUser)

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
