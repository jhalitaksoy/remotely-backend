package main

import (
	"log"
	"testing"
)

func TestRoomProvider(t *testing.T) {
	myContext := newMyContextForTest()

	roomDB, err := myContext.RoomStore.Create(RoomDB{
		Name:    "TestRoom1",
		OwnerID: 0,
	})

	if err != nil {
		log.Println(err)
		t.Fail()
	}

	room, err := myContext.RoomProvider.GetFromCache(roomDB.ID)
	if err != nil {
		log.Println(err)
		t.Fail()
	}

	room2, err := myContext.RoomProvider.GetFromCache(roomDB.ID)
	if err != nil {
		log.Println(err)
		t.Fail()
	}

	if room.Name != room2.Name {
		log.Println("Room names not equal")
		t.Fail()
	}
}

func TestRoomProviderWithDatabase(t *testing.T) {
	LoadEnviromentVariables()
	myContext := newMyContext()
	status := myContext.AuthService.Register(&RegisterParameters{
		Name:     "hlt",
		Password: "1234",
	})

	if status >= 400 {
		log.Printf("Register result with %d", status)
		t.Fail()
		return
	}

	status, jwt := myContext.AuthService.Login(&LoginParameters{
		Name:     "hlt",
		Password: "1234",
	})

	if status >= 400 {
		log.Printf("Register result with %d", status)
		t.Fail()
		return
	}

	log.Printf("JWT key is %s", jwt)
}
