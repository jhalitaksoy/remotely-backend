package main

import (
	"errors"
	"io"
	"log"

	"github.com/pion/webrtc/v3"
)

//MediaRoom is
type MediaRoom struct {
	RoomID          int
	API             *webrtc.API
	UserAudioTracks map[int]*webrtc.TrackLocalStaticRTP
	VideoTrack      *webrtc.TrackLocalStaticRTP
}

//NewMediaRoom is
func NewMediaRoom(id int) *MediaRoom {
	mediaEngine := NewMediaEngine()
	api := NewAPI(mediaEngine)

	//userAudioTracks := CreateUserAudioTracks()
	videoTrack := NewVideoTrack()

	return &MediaRoom{
		RoomID:          id,
		API:             api,
		UserAudioTracks: make(map[int]*webrtc.TrackLocalStaticRTP),
		VideoTrack:      videoTrack,
	}
}

//NewPeerConnection is
func (mediaRoom *MediaRoom) NewPeerConnection() *webrtc.PeerConnection {
	return NewPeerConnection(mediaRoom.API)
}

func (mediaRoom *MediaRoom) findSuitableAudioTrack(user *User) *webrtc.TrackLocalStaticRTP {
	audioTrack, ok := mediaRoom.UserAudioTracks[user.ID]
	if ok {
		log.Println("Founded old audio track")
		return audioTrack
	} else {
		audioTrack = NewAudioTrack()
		mediaRoom.UserAudioTracks[user.ID] = audioTrack
	}
	return audioTrack
}

func (mediaRoom *MediaRoom) findAudioTrackByUser(user *User) *webrtc.TrackLocalStaticRTP {
	audioTrack := mediaRoom.UserAudioTracks[user.ID]
	return audioTrack
}

func (mediaRoom *MediaRoom) RemoveAudioTrackByUser(user *User) {
	delete(mediaRoom.UserAudioTracks, user.ID)
}

//JoinUser is
func (mediaRoom *MediaRoom) JoinUser(
	peer *Peer,
	sd webrtc.SessionDescription) (*webrtc.SessionDescription, error) {

	pc := peer.PeerConnection

	userAudioTrack := mediaRoom.findSuitableAudioTrack(peer.User)
	if userAudioTrack == nil {
		print("Room is full!")
		return nil, errors.New("Room is full")
	}

	mediaRoom.allowSendAudioTrackToAllPeers(peer)

	mediaRoom.allowSendAudioTrackToAllPeers2(peer, userAudioTrack)

	AllowReceiveAudioTrack(pc)

	if peer.IsPublisher {
		log.Println("Publisher")
		AllowReceiveVideoTrack(pc)

	} else {
		log.Println("Client")
		AllowSendVideoTrack(pc, mediaRoom.VideoTrack)
	}

	mediaRoom.addOnTrack(peer)

	log.Printf("Added RoomUser id : %d. RoomUser Count : %d", peer.User.ID, len(peer.Room.Users))

	DataChannelHandler(myContext, peer)

	pc.OnNegotiationNeeded(func() {
		log.Println("OnNegotiationNeeded")
	})

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

	peer.Room.UpdateSDPs(myContext, peer)
	peer.Room.ListenIceMessages(peer, myContext)

	return &answer, nil
}

func (mediaRoom *MediaRoom) allowSendAudioTrackToAllPeers(peer *Peer) {
	for userID, userAudioTrack := range mediaRoom.UserAudioTracks {
		if peer.User.ID != userID {
			AllowSendAudioTrack(peer.PeerConnection, userAudioTrack)
		}
	}
}

func (mediaRoom *MediaRoom) allowSendAudioTrackToAllPeers2(peer *Peer, remoteAudioTrack *webrtc.TrackLocalStaticRTP) {
	for _, roomUser := range peer.Room.Users {
		if roomUser.User.ID != peer.User.ID {
			AllowSendAudioTrack(roomUser.PeerConnection, remoteAudioTrack)
		}
	}
}

func (mediaRoom *MediaRoom) addOnTrack(roomUser *Peer) {
	// Set a handler for when a new remote track starts, this just distributes all our packets
	// to connected peers
	roomUser.PeerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Println("Track acquired", remoteTrack.Kind(), remoteTrack.Codec())

		mediaRoom.onTrack(roomUser, remoteTrack, receiver)
	})
}

func (mediaRoom *MediaRoom) onTrack(roomUser *Peer, remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
	// This can be less wasteful by processing incoming RTCP events, then we would emit a NACK/PLI when a viewer requests it
	sendPLIInterval(roomUser.PeerConnection, remoteTrack)

	log.Printf("Sending New Track %s", remoteTrack.Kind())

	rtpBuf := make([]byte, 1400)
	for {
		i, _, err := remoteTrack.Read(rtpBuf)
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

		audioTrack := userAudioTrack

		if audioTrack == nil {
			log.Println("Audio track was nil!")
			break
		}

		if remoteTrack.Kind() == webrtc.RTPCodecTypeAudio {
			if _, err := audioTrack.Write(rtpBuf[:i]); err != nil && err != io.ErrClosedPipe {
				log.Panic(err)
			}
		} else if remoteTrack.Kind() == webrtc.RTPCodecTypeVideo {
			if _, err := mediaRoom.VideoTrack.Write(rtpBuf[:i]); err != nil && err != io.ErrClosedPipe {
				log.Panic(err)
			}
		}
	}
}

//AddStreamer is
func (mediaRoom *MediaRoom) AddStreamer() {

}

//AddViewer is
func (mediaRoom *MediaRoom) AddViewer() {

}
