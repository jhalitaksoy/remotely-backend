package main

type MessageChannel string

type Message []byte

type PeerID = int

type OnMessage = func(*Peer, MessageChannel, Message)

type OnConnectionChange = func(*Peer)

type RealtimeMessageTransport interface {
	Send(PeerID, MessageChannel, Message) error
	Listen(MessageChannel, OnMessage) error

	OnConnected(OnConnectionChange)
	OnDisConnected(OnConnectionChange)
}
