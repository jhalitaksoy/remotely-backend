package main

import (
	"log"

	"github.com/pion/webrtc/v3"
)

//DataChannelHandler is
func DataChannelHandler(myContext *MyContext, peer *Peer) {
	pc := peer.PeerConnection
	pc.OnDataChannel(func(d *webrtc.DataChannel) {
		log.Printf("New DataChannel %s - %d\n", d.Label(), d.ID())
		peer.setDataChannel(d)

		rtmtDataChannel, ok := myContext.RTMT.(*RealtimeMessageTransportDataChannel)
		if !ok {
			log.Println("rtmt is not RealtimeMessageTransportDataChannel")

		} else {
			rtmtDataChannel.AddPeer(peer.User.ID, peer)
		}
	})
}
