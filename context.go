package main

type Context struct {
	User        *User
	Room        *Room
	MediaRoom   *MediaRoom
	RoomUser    *RoomUser
	IsPublisher bool
}

func NewContext(room *Room, mediaRoom *MediaRoom, roomUser *RoomUser, isPublisher bool) *Context {
	return &Context{
		User:        roomUser.User,
		RoomUser:    roomUser,
		Room:        room,
		MediaRoom:   mediaRoom,
		IsPublisher: isPublisher,
	}
}
