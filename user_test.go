package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestAddUser(t *testing.T) {
	for i := 0; i < 10; i++ {
		userTemp := User{Name: "Alice"}
		userRepository.AddUser(&userTemp)
		if userTemp.ID != i {
			t.Fatalf("Expected user id %d, got %d", i, userTemp.ID)
		}
		usersLen := len(userRepository.(*UserRepositoryMock).users)
		if usersLen != i+1 {
			t.Fatalf("Expected users lenght %d, got %d", i+1, usersLen)
		}
	}
}

func TestAddUserHttp(t *testing.T) {
	ts := httptest.NewServer(StartGin())
	defer ts.Close()

	userName := "Alice"
	url := fmt.Sprintf("%s/user/login/%s", ts.URL, userName)

	resp, err := http.Post(
		url,
		"",
		bytes.NewBufferString(""),
	)

	if err != nil {
		t.Fatalf("Excpected no error %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %v", resp.StatusCode)
	} else {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		bodyString := string(bodyBytes)
		bodyInt, err := strconv.Atoi(bodyString)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("User id is %d", bodyInt)
		fmt.Println("")
	}
}
