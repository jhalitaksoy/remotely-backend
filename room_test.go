package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestRoom(t *testing.T) {
	user1 := &User{Name: "Alice", Password: "1234"}
	user2 := &User{Name: "Bob", Password: "1234"}
	err := userRepository.RegisterUser(user1)
	if err != nil {
		t.Fatal(err)
	}
	err = userRepository.RegisterUser(user2)
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(StartGin())
	defer ts.Close()

	//Create Room

	room1 := &Room{Name: "Room1"}
	url := fmt.Sprintf("%s/room/create", ts.URL)
	jsonRoom, err := json.Marshal(room1)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := PostWithUser(url, bytes.NewBuffer(jsonRoom), user1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %v", resp.StatusCode)
	}

	//Join Room
	url = fmt.Sprintf("%s/room/join/%d", ts.URL, room1.ID)
	resp, err = PostWithUser(url, bytes.NewBufferString(""), user2)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %v", resp.StatusCode)
	}

	//List Room
	url = fmt.Sprintf("%s/room/listRooms", ts.URL)
	resp, err = PostWithUser(url, bytes.NewBufferString(""), user2)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %v", resp.StatusCode)
	}

	rooms := make([]*Room, 0)
	json.NewDecoder(resp.Body).Decode(&rooms)

	if len(rooms) != 1 {
		t.Fatalf("Expected lenght of rooms 1, got %d", len(rooms))
	}

	//List Room 2
	url = fmt.Sprintf("%s/room/listRooms", ts.URL)
	resp, err = PostWithUser(url, bytes.NewBufferString(""), user1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %v", resp.StatusCode)
	}

	rooms = make([]*Room, 0)
	err = json.NewDecoder(resp.Body).Decode(&rooms)
	if err != nil {
		t.Fatal(err)
	}
	if len(rooms) != 1 {
		t.Fatalf("Expected lenght of rooms 1, got %d", len(rooms))
	}

	//Get Room
	url = fmt.Sprintf("%s/room/get/%d", ts.URL, room1.ID)
	resp, err = PostWithUser(url, bytes.NewBufferString(""), user1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %v", resp.StatusCode)
	}

	room := &Room{}
	err = json.NewDecoder(resp.Body).Decode(room)
	if err != nil {
		t.Fatal(err)
	}
}

func PostWithUser(url string, body io.Reader, user *User) (*http.Response, error) {
	header := http.Header{}
	header.Add("userid", strconv.Itoa(user.ID))
	return PostWithHeader(url, body, header)
}

func PostWithHeader(url string, body io.Reader, header http.Header) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		panic(err)
	}
	req.Header = header
	req.Header.Set("Content-Type", appJSON)
	return client.Do(req)
}
