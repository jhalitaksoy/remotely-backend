package main

//Room is
type Room struct {
	ID    int
	Name  string
	Owner *User
	Users []*RoomUser
	ChatMessages []*ChatMessage
}

//RoomRepository is
type RoomRepository interface {
	GetRoomByID(ID int) *Room
	CreateRoom(*User, *Room) bool
	JoinRoom(*User, *Room) bool
	ListRooms(*User) []*Room
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

var roomRepository RoomRepository = &RoomRepositoryMock{lastRoomID: -1, userRoomsTable: make(map[*User][]*Room)}

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
	room.Owner = user
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
