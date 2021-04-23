package main

import (
	"crypto/rand"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func hashAndSaltPassword(password string) (string, error) {
	bytes, error := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if error != nil {
		return "", error
	}
	hash := string(bytes)
	return hash, nil
}

func createUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid
}

func checkStringsIsEmpty(strs ...string) bool {
	for _, str := range strs {
		if checkStringIsEmpty(str) {
			return true
		}
	}
	return false
}

func checkStringIsEmpty(str string) bool {
	return len(str) == 0
}
