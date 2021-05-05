package main

import (
	"errors"
	"log"

	"github.com/pion/webrtc/v3"
)

//Room is
type Room struct {
	ID           int
	Name         string
	OwnerID      int
	Users        []*Peer
	MediaRoom    *MediaRoom
	ChatMessages []*ChatMessage
	Surveys      []*Survey
	lastSurveyID int
}

type RoomDB struct {
	TableName struct{} `sql:"rooms"`
	ID        int
	Name      string
	OwnerID   int
}

func NewRoom(roomdb RoomDB) *Room {
	return &Room{
		ID:           roomdb.ID,
		Name:         roomdb.Name,
		OwnerID:      roomdb.OwnerID,
		Users:        make([]*Peer, 0),
		MediaRoom:    NewMediaRoom(roomdb.ID),
		ChatMessages: []*ChatMessage{},
		Surveys:      make([]*Survey, 0),
		lastSurveyID: -1,
	}
}

//CreateNewSurvey create new Survey from given survey
func (room *Room) CreateNewSurvey(survey *Survey) *Survey {
	room.lastSurveyID += 1

	for i, option := range survey.Options {
		option.ID = i
	}

	return &Survey{
		ID:               room.lastSurveyID,
		Text:             survey.Text,
		Owner:            survey.Owner,
		Options:          survey.Options,
		Votes:            map[int]*SurveyOption{},
		ParticipantCount: 0,
	}
}

//VoteSurvey ...
func (room *Room) VoteSurvey(id int, user *User) {

}

func (room *Room) JoinUserToRoom(myContext *MyContext, user *User, sd webrtc.SessionDescription, isPublisher bool) (*webrtc.SessionDescription, error) {
	mediaRoom := room.MediaRoom
	if mediaRoom == nil {
		return nil, errors.New("media room not found")
	}

	pc := mediaRoom.NewPeerConnection()
	peer := NewPeer(user, pc, room, isPublisher)
	room.addPeer(peer)

	myContext.RoomProviderGC.OnUserConnectionOpen(peer)

	pc.OnConnectionStateChange(func(pcs webrtc.PeerConnectionState) {
		log.Println(pcs.String())
		room.onPeerConnectionChange(peer, myContext, pcs)
	})

	ListenMessages(myContext, peer)

	return mediaRoom.JoinUser(peer, sd)
}

func (room *Room) onPeerConnectionChange(peer *Peer, myContext *MyContext, pcs webrtc.PeerConnectionState) {
	switch pcs {
	case webrtc.PeerConnectionStateNew:
	case webrtc.PeerConnectionStateConnecting:
	case webrtc.PeerConnectionStateConnected:
	case webrtc.PeerConnectionStateFailed:
		//todo look later
		//myContext.RoomProviderGC.OnUserConnectionClose(context)
	case webrtc.PeerConnectionStateClosed:
		//todo look later
		//myContext.RoomProviderGC.OnUserConnectionClose(context)
	case webrtc.PeerConnectionStateDisconnected:
		myContext.RoomProviderGC.OnUserConnectionClose(peer)
	}
}

func (room *Room) addPeer(peer *Peer) {
	for i, eachRoomUser := range room.Users {
		if eachRoomUser.User.ID == peer.User.ID {
			//refactor
			len := len(room.Users) - 1
			room.Users[i] = room.Users[len]
			room.Users[len] = nil
			room.Users = room.Users[:len]
			break
		}
	}
	room.Users = append(room.Users, peer)
}

func (room *Room) addChatMessage(chatMessage *ChatMessage) {
	room.ChatMessages = append(room.ChatMessages, chatMessage)
}

func (room *Room) addSurvey(survey *Survey) {
	room.Surveys = append(room.Surveys, survey)
}

func (room *Room) removeSurvey(survey *Survey) {
	for i, _survey := range room.Surveys {
		if _survey == survey {
			room.removeSurveyAt(i)
			break
		}
	}
}

func (room *Room) removeSurveyAt(s int) {
	room.Surveys = append(room.Surveys[:s], room.Surveys[s+1:]...)
}

func (room *Room) getSurveyByID(id int) *Survey {
	for _, survey := range room.Surveys {
		if survey.ID == id {
			return survey
		}
	}
	return nil
}

func (room *Room) RemoveRoomUser(roomUser *Peer) bool {
	room.MediaRoom.RemoveAudioTrackByUser(roomUser.User)
	for i, eachRoomUser := range room.Users {
		if eachRoomUser.User.ID == roomUser.User.ID {
			room.Users = removeRoomUserByIndex(room.Users, i)
			return true
		}
	}
	return false
}
func (room *Room) MustRemove() bool {
	return len(room.Users) == 0
}

func removeRoomUserByIndex(s []*Peer, i int) []*Peer {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
