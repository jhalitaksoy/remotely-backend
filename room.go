package main

import (
	"errors"
	"log"

	"github.com/pion/webrtc/v2"
)

//Room is
type Room struct {
	ID           int
	Name         string
	OwnerID      int
	Users        []*RoomUser
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
		lastSurveyID: -1,
		Users:        make([]*RoomUser, 0),
		ChatMessages: []*ChatMessage{},
		Surveys:      make([]*Survey, 0),
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

func (room *Room) JoinUserToRoom(user *User, sd webrtc.SessionDescription, isPublisher bool) (*webrtc.SessionDescription, error) {
	mediaRoom := mediaRoomRepository.GetMediaRoomByRoomID(room.ID)
	if mediaRoom == nil {
		return nil, errors.New("media room not found")
	}

	pc := mediaRoom.NewPeerConnection()
	roomUser := NewRoomUser(user, pc)
	room.addRoomUser(roomUser)

	context := NewContext(room, mediaRoom, roomUser, isPublisher)

	pc.OnConnectionStateChange(func(pcs webrtc.PeerConnectionState) {
		log.Println(pcs.String())
		OnPeerConnectionChange(context, pcs)
	})

	return mediaRoom.JoinUser(context, sd)
}

func (room *Room) addRoomUser(newRoomUser *RoomUser) {
	for i, eachRoomUser := range room.Users {
		if eachRoomUser.User.ID == newRoomUser.User.ID {
			//refactor
			len := len(room.Users) - 1
			room.Users[i] = room.Users[len]
			room.Users[len] = nil
			room.Users = room.Users[:len]
			break
		}
	}
	room.Users = append(room.Users, newRoomUser)
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

func (room *Room) RemoveRoomUser(roomUser *RoomUser) bool {
	for i, eachRoomUser := range room.Users {
		if eachRoomUser.User.ID == roomUser.User.ID {
			room.Users = removeRoomUserByIndex(room.Users, i)
			return true
		}
	}
	return false
}

func removeRoomUserByIndex(s []*RoomUser, i int) []*RoomUser {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

//RoomRepository is
type RoomRepository interface {
	GetRoomByID(ID int) *Room
	CreateRoom(*User, *Room) bool
	JoinRoom(*User, *Room) bool
	ListRooms(*User) []*Room
}

//var roomRepository RoomRepository = &RoomRepositoryMock{lastRoomID: -1, userRoomsTable: make(map[*User][]*Room)}

// RoomRepositoryMock is
type RoomRepositoryMock struct {
	rooms          []*Room
	userRoomsTable map[*User][]*Room
	lastRoomID     int
}

//CreateRoom is
func (repo *RoomRepositoryMock) CreateRoom(user *User, room *Room) bool {
	repo.lastRoomID = repo.lastRoomID + 1
	room.ID = repo.lastRoomID
	room.OwnerID = user.ID
	room.lastSurveyID = 0
	repo.JoinRoom(user, room)
	repo.rooms = append(repo.rooms, room)
	return true
}

//JoinRoom is
func (repo *RoomRepositoryMock) JoinRoom(user *User, room *Room) bool {
	list := repo.userRoomsTable[user]
	for _, eachRoom := range list {
		if eachRoom.ID == room.ID {
			return false
		}
	}
	repo.userRoomsTable[user] = append(list, room)
	return true
}

//ListRooms is
func (repo *RoomRepositoryMock) ListRooms(user *User) []*Room {
	list := repo.userRoomsTable[user]
	if list == nil {
		return make([]*Room, 0)
	}
	return list
}

//GetRoomByID is
func (repo *RoomRepositoryMock) GetRoomByID(ID int) *Room {
	for _, room := range repo.rooms {
		if room.ID == ID {
			return room
		}
	}
	return nil
}
