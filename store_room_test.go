package main

import (
	"log"
	"testing"
)

func TestRoomStoreDB(t *testing.T) {
	LoadEnviromentVariables()
	myContext := newMyContext()

	roomDB, err := myContext.RoomStore.Create(RoomDB{
		Name:    "TestRoom1",
		OwnerID: 2,
	})

	if err != nil {
		log.Println(err)
		t.Fail()
	}

	room, err := myContext.RoomStore.GetByID(roomDB.ID)
	if err != nil {
		log.Println(err)
		t.Fail()
	}

	err = myContext.RoomStore.Delete(room.ID)
	if err != nil {
		log.Println(err)
		t.Fail()
	}

	_, err = myContext.RoomStore.GetByID(room.ID)
	if err == nil {
		log.Println("Room cannot deleted")
		t.Fail()
	}
}

func TestRoomStoreDBListRoomsByUser(t *testing.T) {
	LoadEnviromentVariables()
	myContext := newMyContext()
	userId := 2
	_, err := myContext.RoomStore.Create(RoomDB{
		Name:    "TestRoom1",
		OwnerID: userId,
	})

	if err != nil {
		log.Println(err)
		t.Fail()
	}

	_, err = myContext.RoomStore.Create(RoomDB{
		Name:    "TestRoom2",
		OwnerID: 2,
	})

	if err != nil {
		log.Println(err)
		t.Fail()
	}

	rooms, err := myContext.RoomStore.GetByUserID(userId)
	if err != nil {
		log.Println(err)
		t.Fail()
	}

	if len(rooms) != 2 {
		log.Println("Rooms lenght is not match")
		t.Fail()
	}
}
