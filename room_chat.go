package main

import (
	"encoding/json"
	"log"
)

type DataChannelUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ChatMessage struct {
	Text string          `json:"text"`
	User DataChannelUser `json:"user"`
}

type ChatMessageData struct {
	UserID   string
	UserName string
	Message  string
}

func OnChatMessage(myContext *MyContext, peer *Peer, message Message) {
	var _chatMessage ChatMessage
	err := json.Unmarshal(message, &_chatMessage)
	if err != nil {
		log.Printf("(OnChatMessage) Error %v", err)
	}
	user := DataChannelUser{
		ID:   peer.User.ID,
		Name: peer.User.Name,
	}
	chatMessage := &ChatMessage{
		User: user,
		Text: _chatMessage.Text,
	}
	SendChatMessage(myContext, peer.Room, chatMessage)
}

func SendChatMessage(myContext *MyContext, room *Room, message *ChatMessage) {
	user := myContext.UserStore.GeById(message.User.ID)
	if user == nil {
		log.Fatalf("User not found! id : %d", message.User.ID)
		return
	}
	message.User.Name = user.Name

	room.addChatMessage(message)

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("An error occured when converting chat message to json %v", err)
		return
	}

	for _, peer := range room.Users {
		myContext.RTMT.Send(peer.User.ID, ChannelChat, data)
	}
}
