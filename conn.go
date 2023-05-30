package main

import (
	"os"

	"github.com/go-pg/pg/v10"
	_ "github.com/lib/pq"
)

type UserDBconfig struct {
	conn *pg.DB
}

func Open() *pg.DB {
	db := UserDBconfig{}
	db.conn = pg.Connect(&pg.Options{
		Addr:     ":5432",
		User:     "postgres",
		Password: os.Getenv("PG_PWD"),
		Database: "bookserver",
	})
	return db.conn
}
