package main

type RoomProvider interface {
	GetFromCache(roomID int) (*Room, error)
	RemoveFromCache(roomID int)
	SetRoomStore(RoomStore)
	GetRoomStore() RoomStore
}

type RoomProviderImpl struct {
	dbStore RoomStore
	Rooms   map[int]*Room
}

func NewRoomProviderImpl(dbStore RoomStore) *RoomProviderImpl {
	roomProvider := RoomProviderImpl{
		dbStore: dbStore,
		Rooms:   make(map[int]*Room),
	}
	//todo remove listener in somewhere
	dbStore.SetOnRoomChangeListener(roomProvider.OnRoomChange)

	return &roomProvider
}

func (roomProvider *RoomProviderImpl) SetRoomStore(roomStore RoomStore) {
	roomProvider.dbStore = roomStore
}

func (roomProvider *RoomProviderImpl) GetRoomStore() RoomStore {
	return roomProvider.dbStore
}

func (roomProvider *RoomProviderImpl) GetFromCache(roomID int) (*Room, error) {
	roomCache := roomProvider.Rooms[roomID]
	if roomCache == nil {
		room, err := roomProvider.StoreRoomInCache(roomID)
		if err != nil {
			return nil, err
		}
		roomCache = room
	}

	return roomCache, nil
}

func (roomProvider *RoomProviderImpl) RemoveFromCache(roomID int) {
	delete(roomProvider.Rooms, roomID)
}

func (roomProvider *RoomProviderImpl) StoreRoomInCache(roomID int) (*Room, error) {
	roomdb, err := roomProvider.GetRoomStore().GetByID(roomID)
	if err != nil {
		return nil, err
	}

	room := NewRoom(*roomdb)
	roomProvider.Rooms[roomID] = room
	return room, nil
}

func (roomProvider *RoomProviderImpl) OnRoomChange(roomID int) {
	roomProvider.RemoveFromCache(roomID) //todo look later
}
