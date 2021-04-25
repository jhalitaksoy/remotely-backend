package main

import "log"

type RoomProviderGC interface {
	OnUserConnectionOpen(*Context)
	OnUserConnectionClose(*Context)
	SetRoomProvider(RoomProvider)
	GetRoomProvider() RoomProvider
	SetUserStore(UserStore)
	GetUserStore() UserStore
}

type UserConnectionCout struct {
	OpenCount  int
	CloseCount int
}

func (userConnectionCount *UserConnectionCout) MustRemove() bool {
	return userConnectionCount.OpenCount <= userConnectionCount.CloseCount
}

func NewUserConnectionCout() *UserConnectionCout {
	return &UserConnectionCout{
		OpenCount:  0,
		CloseCount: 0,
	}
}

type RoomProviderGCImpl struct {
	roomProvider           RoomProvider
	userStore              UserStore
	userConnectionCountMap map[int]*UserConnectionCout
}

func NewRoomProviderGCImpl(roomProvider RoomProvider, userStore UserStore) *RoomProviderGCImpl {
	return &RoomProviderGCImpl{
		roomProvider:           roomProvider,
		userStore:              userStore,
		userConnectionCountMap: make(map[int]*UserConnectionCout),
	}
}

func (roomProviderGC *RoomProviderGCImpl) SetRoomProvider(roomProvider RoomProvider) {
	roomProviderGC.roomProvider = roomProvider
}

func (roomProviderGC *RoomProviderGCImpl) GetRoomProvider() RoomProvider {
	return roomProviderGC.roomProvider
}

func (roomProviderGC *RoomProviderGCImpl) SetUserStore(userStore UserStore) {
	roomProviderGC.userStore = userStore
}

func (roomProviderGC *RoomProviderGCImpl) GetUserStore() UserStore {
	return roomProviderGC.userStore
}

func (roomProviderGC *RoomProviderGCImpl) OnUserConnectionOpen(context *Context) {
	userConnectionCount := roomProviderGC.getUserConnectionCount(context.User)
	userConnectionCount.OpenCount++
}

func (roomProviderGC *RoomProviderGCImpl) OnUserConnectionClose(context *Context) {
	userConnectionCount := roomProviderGC.getUserConnectionCount(context.User)
	userConnectionCount.CloseCount++
	roomProviderGC.onUserConnectionClose(context, userConnectionCount)
	roomProviderGC.removeRoomIfRequired(context)
}

func (roomProviderGC *RoomProviderGCImpl) getUserConnectionCount(user *User) *UserConnectionCout {
	userConnectionCount := roomProviderGC.userConnectionCountMap[user.ID]
	if userConnectionCount == nil {
		userConnectionCount = NewUserConnectionCout()
		roomProviderGC.userConnectionCountMap[user.ID] = userConnectionCount
	}
	return userConnectionCount
}

func (roomProviderGC *RoomProviderGCImpl) onUserConnectionClose(context *Context, userConnectionCount *UserConnectionCout) {
	if context.User.Anonymous {
		roomProviderGC.userStore.Delete(context.User.ID)
		context.Room.RemoveRoomUser(context.RoomUser)
		log.Printf("Removed anonymous user from room : %v", context.RoomUser.User.ID)
	} else {
		if userConnectionCount.MustRemove() {
			context.Room.RemoveRoomUser(context.RoomUser)
			log.Printf("Removed user from room : %v", context.RoomUser.User.ID)
		}
	}
}

func (roomProviderGC *RoomProviderGCImpl) removeRoomIfRequired(context *Context) {
	if context.Room.MustRemove() {
		roomProviderGC.roomProvider.RemoveFromCache(context.Room.ID)
		log.Printf("Removed room from cache : %v", context.Room.ID)
	}
}
