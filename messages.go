package main

const (
	ChannelChat          = "chat"
	ChannelSurveyCreate  = "survey_create"
	ChannelSurveyDestroy = "survey_destroy"
	ChannelSurveyUpdate  = "survey_update"
	ChannelSurveyVote    = "survey_vote"
)

func ListenMessages(myContext *MyContext, peer *Peer) {
	myContext.RTMT.Listen(ChannelChat, func(peer *Peer, channel MessageChannel, message Message) {
		OnChatMessage(myContext, peer, message)
	})
	myContext.RTMT.Listen(ChannelSurveyCreate, func(peer *Peer, channel MessageChannel, message Message) {
		OnSurveyCreate(myContext, peer, message)
	})
	myContext.RTMT.Listen(ChannelSurveyVote, func(peer *Peer, channel MessageChannel, message Message) {
		OnSurveyVote(myContext, peer, message)
	})
	myContext.RTMT.Listen(ChannelReady, func(p *Peer, mc MessageChannel, m Message) {
		sendOffer(peer, nil, myContext)
	})
}
