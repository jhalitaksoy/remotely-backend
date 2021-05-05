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
			panic("rtmt is not RealtimeMessageTransportDataChannel")
		}

		rtmtDataChannel.AddPeer(peer.User.ID, peer)
	})
}
