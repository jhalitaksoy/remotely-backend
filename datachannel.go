package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/pion/webrtc/v2"
)

//MessageType is
type MessageType int8

const (
	chatMessage MessageType = iota
	surveyCreate
	surveyVote
	surveyUpdate
	surveyEnd
)

//DataChannelUser is
type DataChannelUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

//ChatMessage is
type ChatMessage struct {
	Text string          `json:"text"`
	User DataChannelUser `json:"user"`
}

//ChatMessageData is
type ChatMessageData struct {
	UserID   string
	UserName string
	Message  string
}

//DataChannelHandler is
func DataChannelHandler(pc *webrtc.PeerConnection, room *Room, roomUser *RoomUser) {
	// Register data channel creation handling
	pc.OnDataChannel(func(d *webrtc.DataChannel) {
		log.Printf("New DataChannel %s - %d\n", d.Label(), d.ID())

		roomUser.SetDataChannel(d)

		d.OnOpen(func() {
			log.Printf("Open Data channel %s - %d\n", d.Label(), d.ID())
		})

		d.OnClose(func() {
			log.Printf("Closed Data channel %s - %d.\n", d.Label(), d.ID())
		})

		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			onMessage(room, roomUser, msg)
		})
	})
}

func onMessage(room *Room, roomUser *RoomUser, msg webrtc.DataChannelMessage) {

	len := len(msg.Data)

	messageType := MessageType(msg.Data[0])

	wholeData := make([]byte, len)

	for i := 0; i < len; i++ {
		wholeData[i] = msg.Data[i]
	}

	switch messageType {
	case chatMessage:
		{
			onChatMessage(room, roomUser, msg.Data[1:len])
		}
	case surveyCreate:
		{
			onSurveyCreate(room, roomUser, msg.Data[1:len])
		}
	case surveyVote:
		{
			onSurveyVote(room, roomUser, msg.Data[1:len])
		}
	}
}

func sendMessage(dataChannel *webrtc.DataChannel, messageType MessageType, message interface{}) {
	strMessage, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}

	messageTypeLength := 1
	messageLenght := len(strMessage)
	length := messageTypeLength + messageLenght

	bytes := make([]byte, length)

	bytes[0] = byte(messageType)

	for i, char := range strMessage {
		bytes[i+messageTypeLength] = char
	}

	dataChannel.Send(bytes)
}

func onChatMessage(room *Room, roomUser *RoomUser, data []byte) {
	var message ChatMessage
	err := json.Unmarshal(data, &message)
	if err != nil {
		panic(err)
	}
	SendChatMessage(room, &ChatMessage{
		User: DataChannelUser{
			ID:   roomUser.User.ID,
			Name: roomUser.User.Name,
		},
		Text: message.Text,
	})
}

//SendChatMessage is
func SendChatMessage(room *Room, message *ChatMessage) {
	user := userRepository.GetUserByID(message.User.ID)
	if user == nil {
		log.Fatalf("User not found! id : %d", message.User.ID)
		return
	}

	message.User.Name = user.Name

	room.addChatMessage(message)

	for _, user := range room.Users {
		sendMessage(user.DataChannel, chatMessage, message)
	}
}

type surveyCreateParams struct {
	Text    string          `json:"text"`
	Options []*SurveyOption `json:"options"`
}

func onSurveyCreate(room *Room, roomUser *RoomUser, data []byte) {
	var surveyCreateParams surveyCreateParams
	err := json.Unmarshal(data, &surveyCreateParams)
	if err != nil {
		panic(err)
	}

	surveyFromUser := &Survey{
		Text:    surveyCreateParams.Text,
		Options: surveyCreateParams.Options,
	}

	survey := room.CreateNewSurvey(surveyFromUser)
	room.addSurvey(survey)

	sendNewSurveyToRoom(room, survey)
	log.Println("On New Survey")

	go func() {
		time.Sleep(surveyEndDuration)
		log.Println("On Survey Destroy! Survey ID : ", survey.ID)

		surveyEndMessage := map[string]interface{}{
			"surveyID": survey.ID,
		}

		room.removeSurvey(survey)

		sendMessageToRoom(room, surveyEnd, surveyEndMessage)
	}()
}

func sendNewSurveyToRoom(room *Room, survey *Survey) {
	sendMessageToRoom(room, surveyCreate, survey)
}

func onSurveyVote(room *Room, roomUser *RoomUser, data []byte) {

	var vote Vote
	err := json.Unmarshal(data, &vote)
	if err != nil {
		panic(err)
	}

	log.Printf("On Vote Survey. Room : %d, User : %d, Survey : %d, Vote : %d",
		room.ID, roomUser.User.ID, vote.SurveyID, vote.OptionID)

	survey := room.getSurveyByID(vote.SurveyID)

	if survey != nil {
		survey.Vote(roomUser.User, vote.OptionID)
		sendSurveyUpdatedToRoom(room, survey)
	} else {
		log.Printf("Survey Not Found! Room : %d, User : %d, Survey : %d, Vote : %d",
			room.ID, roomUser.User.ID, vote.SurveyID, vote.OptionID)
	}
}

func sendSurveyUpdatedToRoom(room *Room, survey *Survey) {
	sendMessageToRoom(room, surveyUpdate, survey)
}

func sendMessageToRoom(room *Room, messageType MessageType, message interface{}) {
	strMessage, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}

	messageTypeLength := 1
	messageLenght := len(strMessage)
	length := messageTypeLength + messageLenght

	bytes := make([]byte, length)

	bytes[0] = byte(messageType)

	for i, char := range strMessage {
		bytes[i+messageTypeLength] = char
	}

	for _, user := range room.Users {
		user.DataChannel.Send(bytes)
	}
}
