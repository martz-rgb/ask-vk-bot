package main

import (
	"encoding/csv"
	"io"
	"os"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	sync.Mutex
	sql *sqlx.DB

	// TO-DO: some sqlbuilder would be nice
}

func NewDB(connection string) (*DB, error) {
	db, err := sqlx.Open("sqlite3", connection)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{
		sql: db,
	}, nil
}

func (db *DB) Init(filename string) error {
	schema, err := os.Open(filename)
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

func (db *DB) LoadCsv(name string) error {
	file, err := os.Open(name)
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
		_, err := tx.Stmt(stmt).Exec(record[0], record[1], record[2], record[3])
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
