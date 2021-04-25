package main

type OnRoomChangeEvent = func(roomID int)

type RoomStore interface {
	Create(RoomDB) (*RoomDB, error)
	GetByID(int) (*RoomDB, error)
	GetByName(string) ([]*RoomDB, error)
	GetByUserID(int) ([]*RoomDB, error)
	Update(*RoomDB) error
	Delete(int) error
	SetOnRoomChangeListener(event OnRoomChangeEvent)
	RemoveOnRoomChangeListener()
}

type RoomStoreImpl struct {
	Rooms             map[int]*RoomDB
	lastRoomID        int
	onRoomChangeEvent OnRoomChangeEvent
}

func NewRoomStoreImpl() *RoomStoreImpl {
	return &RoomStoreImpl{
		Rooms:      make(map[int]*RoomDB),
		lastRoomID: -1,
	}
}

func (roomStore *RoomStoreImpl) Create(room RoomDB) (*RoomDB, error) {
	room.ID = roomStore.createNewRoomId()
	roomStore.Rooms[room.ID] = &room
	return &room, nil
}

func (roomStore *RoomStoreImpl) GetByID(id int) (*RoomDB, error) {
	return roomStore.Rooms[id], nil
}

func (roomStore *RoomStoreImpl) GetByName(name string) ([]*RoomDB, error) {
	rooms := make([]*RoomDB, 0)
	for _, room := range roomStore.Rooms {
		if room.Name == name {
			rooms = append(rooms, room)
		}
	}
	return rooms, nil
}

func (roomStore *RoomStoreImpl) GetByUserID(userID int) ([]*RoomDB, error) {
	rooms := make([]*RoomDB, 0)
	for _, room := range roomStore.Rooms {
		if room.OwnerID == userID {
			rooms = append(rooms, room)
		}
	}
	return rooms, nil
}

func (roomStore *RoomStoreImpl) Update(room *RoomDB) error {
	roomStore.Rooms[room.ID] = room
	roomStore.fireOnRoomChange(room.ID)
	return nil
}

func (roomStore *RoomStoreImpl) Delete(id int) error {
	roomStore.Rooms[id] = nil
	roomStore.fireOnRoomChange(id)
	return nil
}

func (roomStore *RoomStoreImpl) createNewRoomId() int {
	roomStore.lastRoomID++
	return roomStore.lastRoomID
}

func (roomStore *RoomStoreImpl) SetOnRoomChangeListener(onRoomChangeEvent OnRoomChangeEvent) {
	roomStore.onRoomChangeEvent = onRoomChangeEvent
}
func (roomStore *RoomStoreImpl) RemoveOnRoomChangeListener() {
	roomStore.onRoomChangeEvent = nil
}

func (roomStore *RoomStoreImpl) fireOnRoomChange(roomID int) {
	if roomStore.onRoomChangeEvent != nil {
		roomStore.onRoomChangeEvent(roomID)
	}
}
