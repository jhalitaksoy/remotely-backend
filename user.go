package main

import (
	"github.com/pion/webrtc/v3"
)

//User is
type User struct {
	ID        int
	Name      string
	Anonymous bool
}

//Peer is
type Peer struct {
	User           *User
	PeerConnection *webrtc.PeerConnection
	Room           *Room
	DataChannel    *webrtc.DataChannel
	IsPublisher    bool
}

//NewPeer is
func NewPeer(user *User, peerConnection *webrtc.PeerConnection, room *Room, isPublisher bool) *Peer {
	return &Peer{
		User:           user,
		PeerConnection: peerConnection,
		Room:           room,
		IsPublisher:    isPublisher,
	}
}

func (peer *Peer) setDataChannel(dataChannel *webrtc.DataChannel) {
	peer.DataChannel = dataChannel
}
