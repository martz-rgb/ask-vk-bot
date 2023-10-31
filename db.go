package main

import (
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type Db struct {
	sync.Mutex
	sql *sql.DB
}

func NewDb(connection string) (*Db, error) {
	db, err := sql.Open("sqlite3", connection)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Db{
		sql: db,
	}, nil
}
