package main

import (
	"errors"
	"log"

	"github.com/pion/webrtc/v3"
)

var ErrDataChannelNotFound = errors.New("DataChannel not found")
var ErrAlreadyListening = errors.New("already listening")

type RealtimeMessageTransportDataChannel struct {
	MessageChannels map[MessageChannel]OnMessage
	Peers           map[PeerID]*Peer
	onConnected     OnConnectionChange
	onDisConnected  OnConnectionChange
}

func NewRealTimeMessageTransportDataChannel() *RealtimeMessageTransportDataChannel {
	return &RealtimeMessageTransportDataChannel{
		MessageChannels: make(map[MessageChannel]func(*Peer, MessageChannel, Message)),
		Peers:           make(map[int]*Peer),
	}
}

func (rtmt *RealtimeMessageTransportDataChannel) Send(peerID PeerID, channel MessageChannel, message Message) error {
	peer, ok := rtmt.Peers[peerID]
	if ok {
		data := EncodeMessage(channel, message)
		peer.DataChannel.Send(data)
		return nil
	} else {
		return ErrDataChannelNotFound
	}
}

func (rtmt *RealtimeMessageTransportDataChannel) Listen(messageChannel MessageChannel, onMessage OnMessage) error {
	_, ok := rtmt.MessageChannels[messageChannel]
	if ok {
		return ErrAlreadyListening
	}
	rtmt.MessageChannels[messageChannel] = onMessage
	return nil
}

func (rtmt *RealtimeMessageTransportDataChannel) OnConnected(onConnectionChange OnConnectionChange) {
	rtmt.onConnected = onConnectionChange
}

func (rtmt *RealtimeMessageTransportDataChannel) OnDisConnected(onConnectionChange OnConnectionChange) {
	rtmt.onDisConnected = onConnectionChange
}

func (rtmt *RealtimeMessageTransportDataChannel) AddPeer(peerID PeerID, peer *Peer) {
	peer.DataChannel.OnOpen(func() {
		rtmt.OnDataChannelOpen(peerID, peer)
	})

	peer.DataChannel.OnClose(func() {
		rtmt.OnDataChannelClose(peerID, peer)
	})

	peer.DataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		rtmt.OnDataChannelMessage(peer, &msg)
	})
}

func (rtmt *RealtimeMessageTransportDataChannel) OnDataChannelOpen(peerID PeerID, peer *Peer) {
	log.Printf("Open Data channel %s - %d\n", peer.DataChannel.Label(), peer.DataChannel.ID())
	rtmt.Peers[peerID] = peer
	if rtmt.onConnected != nil {
		rtmt.onConnected(peer)
	}
}

func (rtmt *RealtimeMessageTransportDataChannel) OnDataChannelClose(peerID PeerID, peer *Peer) {
	log.Printf("Closed Data channel %s - %d.\n", peer.DataChannel.Label(), peer.DataChannel.ID())
	delete(rtmt.Peers, peerID)
	if rtmt.onDisConnected != nil {
		rtmt.onDisConnected(peer)
	}
}

func (rtmt *RealtimeMessageTransportDataChannel) OnDataChannelMessage(peer *Peer, msg *webrtc.DataChannelMessage) {
	channel, message, err := DecodeMessage(msg.Data)
	if err != nil {
		log.Printf("An error occured when decoding message : %v", err)
		return
	}

	callback, ok := rtmt.MessageChannels[channel]
	if !ok {
		log.Printf("Callback not found for channel : [ %s ]", channel)
		return
	}

	callback(peer, MessageChannel(message), message)
}
