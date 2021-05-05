package main

const (
	ChannelChat = "chat"
)

func ListenMessages(myContext *MyContext, peer *Peer) {
	myContext.RTMT.Listen(ChannelChat, func(peer *Peer, channel MessageChannel, message Message) {
		OnChatMessage(myContext, peer, message)
	})
}
