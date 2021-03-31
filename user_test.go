package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

const (
	appJSON = "application/json"
)

func TestRegisterUser(t *testing.T) {
	userTemp := User{Name: "Alice", Password: "1234"}
	userRepository.RegisterUser(&userTemp)
	if userTemp.ID != 0 {
		t.Fatalf("Expected user id %d, got %d", 0, userTemp.ID)
	}
	usersLen := len(userRepository.(*UserRepositoryMock).users)
	if usersLen != 1 {
		t.Fatalf("Expected users lenght %d, got %d", 1, usersLen)
	}
	userRepository.LoginUser(&userTemp)
	if userTemp.ID != 0 {
		t.Fatalf("Expected user id %d, got %d", 0, userTemp.ID)
	}
}

func TestAddUserHttp(t *testing.T) {
	ts := httptest.NewServer(startGin())
	defer ts.Close()

	user := User{Name: "Alice", Password: "1234"}
	jsonUser, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}

	url := fmt.Sprintf("%s/user/register", ts.URL)

	resp, err := http.Post(
		url,
		appJSON,
		bytes.NewBuffer(jsonUser),
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

	//Login

	url = fmt.Sprintf("%s/user/login", ts.URL)

	resp, err = http.Post(
		url,
		appJSON,
		bytes.NewBuffer(jsonUser),
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
