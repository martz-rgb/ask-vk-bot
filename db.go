package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
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

func (db *Db) Init() error {
	schema, err := os.Open("db/schema.sql")
	if err != nil {
		return err
	}
	defer schema.Close()

	sql, _ := io.ReadAll(schema)

	db.Lock()
	defer db.Unlock()

	_, err = db.sql.Exec(string(sql))
	if err != nil {
		return err
	}

	return nil
}

func (db *Db) LoadCsv() error {
	file, err := os.Open("db/roles.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	content, err := reader.ReadAll()
	if err != nil {
		return err
	}

	// delete header row
	content = content[1:]

	stmt, err := db.sql.Prepare(`INSERT INTO roles 
		(name, tag, shown_name, caption_name)
		VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	tx, err := db.sql.Begin()
	if err != nil {
		return err
	}

	for _, record := range content {
		fmt.Println(record)

		_, err := tx.Stmt(stmt).Exec(record[0], record[1], record[2], record[3])
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
