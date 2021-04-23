package main

import (
	"context"

	"github.com/go-pg/pg"
)

type DataBase struct {
	DB *pg.DB
}

func CreateDatabaseConnection(options *pg.Options) *DataBase {
	db := pg.Connect(options)

	ctx := context.Background()
	_, err := db.ExecContext(ctx, "SELECT 1")
	if err != nil {
		panic(err)
	}
	return &DataBase{
		DB: db,
	}
}
