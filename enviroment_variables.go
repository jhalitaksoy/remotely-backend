package main

import (
	"errors"
	"os"

	"github.com/go-pg/pg"
)

const errorDatabaseConnectionVarsEmpty = "some of the database connection variables is empty"

func getDatabaseConnectionVariables() *pg.Options {
	addr := os.Getenv("DB_ADDR")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	database := os.Getenv("DB_DATABASE")

	if checkStringsIsEmpty(addr, user, password, database) {
		panic(errors.New(errorDatabaseConnectionVarsEmpty))
	}

	options := &pg.Options{
		Addr:     addr,
		User:     user,
		Password: password,
		Database: database,
	}

	return options
}
