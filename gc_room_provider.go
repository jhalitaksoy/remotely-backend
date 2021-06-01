package main

import "log"

type RoomProviderGC interface {
	OnUserConnectionOpen(*Peer)
	OnUserConnectionClose(*Peer)
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
	return userConnectionCount.OpenCount == userConnectionCount.CloseCount
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

func (roomProviderGC *RoomProviderGCImpl) OnUserConnectionOpen(peer *Peer) {
	userConnectionCount := roomProviderGC.getUserConnectionCount(peer.User)
	userConnectionCount.OpenCount++
}

func (roomProviderGC *RoomProviderGCImpl) OnUserConnectionClose(peer *Peer) {
	userConnectionCount := roomProviderGC.getUserConnectionCount(peer.User)
	userConnectionCount.CloseCount++
	roomProviderGC.onUserConnectionClose(peer, userConnectionCount)
	roomProviderGC.removeRoomIfRequired(peer)
}

func (roomProviderGC *RoomProviderGCImpl) getUserConnectionCount(user *User) *UserConnectionCout {
	userConnectionCount := roomProviderGC.userConnectionCountMap[user.ID]
	if userConnectionCount == nil {
		userConnectionCount = NewUserConnectionCout()
		roomProviderGC.userConnectionCountMap[user.ID] = userConnectionCount
	}
	return userConnectionCount
}

func (roomProviderGC *RoomProviderGCImpl) onUserConnectionClose(peer *Peer, userConnectionCount *UserConnectionCout) {
	if peer.User.Anonymous {
		roomProviderGC.userStore.Delete(peer.User.ID)
		peer.Room.RemoveRoomUser(peer)
		log.Printf("Removed anonymous user from room : %v", peer.User.ID)
	} else {
		if userConnectionCount.MustRemove() {
			peer.Room.RemoveRoomUser(peer)
			log.Printf("Removed user from room : %v", peer.User.ID)
		}
	}
}

func (roomProviderGC *RoomProviderGCImpl) removeRoomIfRequired(peer *Peer) {
	if peer.Room.MustRemove() {
		roomProviderGC.roomProvider.RemoveFromCache(peer.Room.ID)
		log.Printf("Removed room from cache : %v", peer.Room.ID)
	}
}
