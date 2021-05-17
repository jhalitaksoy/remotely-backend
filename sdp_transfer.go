package main

import (
	"encoding/json"
	"log"

	"github.com/pion/webrtc/v3"
)

const ChannelSDP = "sdp"
const ChannelICE = "ice"

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

func (room *Room) ListenIceMessages(peer *Peer, myContext *MyContext) {
	room.ListenSelfIceMessage(peer, myContext)
	room.ListenRemoteIceMessage(peer, myContext)
}

func (room *Room) ListenSelfIceMessage(peer *Peer, myContext *MyContext) {
	peer.PeerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		log.Printf("OnICECandidate : %v", i)
		if i != nil {
			json, err := json.Marshal(i.ToJSON())
			if err != nil {
				panic(err)
			}
			myContext.RTMT.Send(peer.User.ID, ChannelICE, Message(json))
		}
	})
}

func (room *Room) ListenRemoteIceMessage(peer *Peer, myContext *MyContext) {
	myContext.RTMT.Listen(ChannelICE, func(p *Peer, mc MessageChannel, m Message) {
		var candidateInit webrtc.ICECandidateInit
		err := json.Unmarshal(m, &candidateInit)
		log.Printf("OnICECandidateMessage : %v", candidateInit)
		if err != nil {
			panic(err)
		}
		p.PeerConnection.AddICECandidate(candidateInit)
	})
}
