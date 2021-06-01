package main

import (
	"log"
	"testing"
)

func TestAuthService(t *testing.T) {
	myContext := newMyContextForTest()
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

func TestAuthServiceWithDatabase(t *testing.T) {
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
