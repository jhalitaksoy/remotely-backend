package main

type RoomStoreDatabaseImpl struct {
	Database          *DataBase
	onRoomChangeEvent OnRoomChangeEvent
}

func NewRoomStoreDatabaseImpl(database *DataBase) *RoomStoreDatabaseImpl {
	return &RoomStoreDatabaseImpl{
		Database: database,
	}
}

func (roomStore *RoomStoreDatabaseImpl) Create(room RoomDB) (*RoomDB, error) {
	_, err := roomStore.Database.DB.Model(&room).Insert()
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (roomStore *RoomStoreDatabaseImpl) GetByID(id int) (*RoomDB, error) {
	roomDB := new(RoomDB)
	err := roomStore.Database.DB.Model(roomDB).Where("id = ?", id).Select()
	return roomDB, err
}

func (roomStore *RoomStoreDatabaseImpl) GetByName(name string) ([]*RoomDB, error) {
	roomDBs := make([]*RoomDB, 0)
	err := roomStore.Database.DB.Model(&roomDBs).Where("name = ?", name).Select()
	return roomDBs, err
}

func (roomStore *RoomStoreDatabaseImpl) GetByUserID(userID int) ([]*RoomDB, error) {
	roomDBs := make([]*RoomDB, 0)
	err := roomStore.Database.DB.Model(&roomDBs).Where("owner_id = ?", userID).Select()
	return roomDBs, err
}

func (roomStore *RoomStoreDatabaseImpl) Update(room *RoomDB) error {
	roomDB := new(RoomDB)
	_, err := roomStore.Database.DB.Model(roomDB).Update()
	if err == nil {
		roomStore.fireOnRoomChange(room.ID)
	}
	return err
}

func (roomStore *RoomStoreDatabaseImpl) Delete(id int) error {
	roomDB := new(RoomDB)
	_, err := roomStore.Database.DB.Model(roomDB).Where("id = ?", id).Delete()
	if err == nil {
		roomStore.fireOnRoomChange(id)
	}
	return err
}

func (roomStore *RoomStoreDatabaseImpl) SetOnRoomChangeListener(onRoomChangeEvent OnRoomChangeEvent) {
	roomStore.onRoomChangeEvent = onRoomChangeEvent
}
func (roomStore *RoomStoreDatabaseImpl) RemoveOnRoomChangeListener() {
	roomStore.onRoomChangeEvent = nil
}

func (roomStore *RoomStoreDatabaseImpl) fireOnRoomChange(roomID int) {
	if roomStore.onRoomChangeEvent != nil {
		roomStore.onRoomChangeEvent(roomID)
	}
}
