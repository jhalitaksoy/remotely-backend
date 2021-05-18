package main

import (
	"encoding/json"
	"log"

	"github.com/pion/webrtc/v3"
)

const ChannelSDPAnswer = "sdp_answer"
const ChannelSDPOffer = "sdp_offer"
const ChannelICE = "ice"

func (room *Room) UpdateSDPs(myContext *MyContext, _peer *Peer) {
	offerOptions := &webrtc.OfferOptions{
		OfferAnswerOptions: webrtc.OfferAnswerOptions{
			VoiceActivityDetection: false,
		},
		ICERestart: false,
	}

	for _, peer := range room.Users {
		if peer.User.ID == _peer.User.ID {
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
		myContext.RTMT.Send(peer.User.ID, ChannelSDPOffer, json)
	}
}

func (room *Room) OnSDPAnswerMessage(peer *Peer, channel MessageChannel, message Message) {
	log.Println("OnSDPAnswerMessage")
	var sdp webrtc.SessionDescription
	err := json.Unmarshal(message, &sdp)
	if err != nil {
		panic(err)
	}
	peer.PeerConnection.SetRemoteDescription(sdp)
}

func (room *Room) OnSDPOfferMessage(peer *Peer, channel MessageChannel, message Message) {
	log.Println("OnSDPOfferMessage")
	var sdp webrtc.SessionDescription
	err := json.Unmarshal(message, &sdp)
	if err != nil {
		panic(err)
	}
	peer.PeerConnection.SetRemoteDescription(sdp)
	answer, err := peer.PeerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}
	peer.PeerConnection.SetLocalDescription(answer)
	answerJson, err := json.Marshal(answer)
	if err != nil {
		panic(err)
	}
	err = myContext.RTMT.Send(peer.User.ID, ChannelSDPAnswer, Message(answerJson))
	if err != nil {
		log.Println(err)
	}
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
