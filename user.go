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

//RoomUser is
type RoomUser struct {
	User           *User
	PeerConnection *webrtc.PeerConnection
	DataChannel    *webrtc.DataChannel
}

//NewRoomUser is
func NewRoomUser(user *User, peerConnection *webrtc.PeerConnection) *RoomUser {
	return &RoomUser{
		User:           user,
		PeerConnection: peerConnection,
	}
}

//SetDataChannel is
func (roomUser *RoomUser) SetDataChannel(dataChannel *webrtc.DataChannel) {
	roomUser.DataChannel = dataChannel
}
