package main

import (
	"encoding/json"

	"github.com/pion/webrtc/v3"
)

const ChannelSDP = "sdp"

func (room *Room) UpdateSDPs(myContext *MyContext, _peer *Peer) {
	offerOptions := &webrtc.OfferOptions{
		OfferAnswerOptions: webrtc.OfferAnswerOptions{
			VoiceActivityDetection: false,
		},
		ICERestart: false,
	}

	myContext.RTMT.Listen(ChannelSDP, room.OnSDPMessage)

	for _, peer := range room.Users {
		if peer.User.ID == _peer.Room.ID {
			continue
		}
		sdp, err := peer.PeerConnection.CreateOffer(offerOptions)
		if err != nil {
			panic(err)
		}
		peer.PeerConnection.SetLocalDescription(sdp)

		json, err := json.Marshal(sdp)
		if err != nil {
			panic(err)
		}
		myContext.RTMT.Send(peer.User.ID, ChannelSDP, json)
	}
}

func (room *Room) OnSDPMessage(peer *Peer, channel MessageChannel, message Message) {
	var sdp webrtc.SessionDescription
	err := json.Unmarshal(message, &sdp)
	if err != nil {
		panic(err)
	}
	peer.PeerConnection.SetRemoteDescription(sdp)
}
