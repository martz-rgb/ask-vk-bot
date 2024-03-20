package db

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/hori-ryota/zaperr"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type DB struct {
	write sync.Mutex
	sql   *sqlx.DB
}

func NewDB(connection string) (*DB, error) {
	// TO-DO: make Sprintf prettier?..
	db, err := sqlx.Open("sqlite3", fmt.Sprintf("%s?_fk=true", connection))
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to open sqlite3 connection",
			zap.String("connection", connection))
	}

	if err := db.Ping(); err != nil {
		return nil, zaperr.Wrap(err, "failed to ping database")
	}

	return &DB{
		sql: db,
	}, nil
}

func (db *DB) Init(filename string, allow_deletion bool) error {
	file, err := os.Open(filename)
	if err != nil {
		return zaperr.Wrap(err, "failed to open file",
			zap.String("filename", filename))
	}

	schema, err := io.ReadAll(file)
	file.Close()
	if err != nil {
		return zaperr.Wrap(err, "failed to  read all from file",
			zap.String("filename", filename))
	}

	err = db.Migrate(string(schema), allow_deletion)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) LoadCsv(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return zaperr.Wrap(err, "failed to open csv file",
			zap.String("filename", name))
	}
	defer file.Close()

	reader := csv.NewReader(file)
	content, err := reader.ReadAll()
	if err != nil {
		return zaperr.Wrap(err, "failed to read all content from csv file")
	}

	// delete header row
	content = content[1:]

	stmt, err := db.sql.Prepare(`INSERT INTO roles 
		(name, tag, shown_name, accusative_name, caption_name)
		VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return zaperr.Wrap(err, "failed to prepare sql statement for csv")
	}

	tx, err := db.sql.Begin()
	if err != nil {
		return zaperr.Wrap(err, "failed to begin sql transaction")
	}

	for _, record := range content {
		_, err := tx.Stmt(stmt).Exec(record[0], record[1], record[2], record[3], record[4])
		if err != nil {
			tx.Rollback()
			return zaperr.Wrap(err, "failed to execute statement in transaction",
				zap.Any("stmt", stmt),
				zap.Any("record", record))
		}
	}

	return tx.Commit()
}

// read-only is concurrency safe
func (db *DB) Select(dest interface{}, query string, args ...interface{}) error {
	return db.sql.Select(dest, query, args...)
}

func (db *DB) Get(dest interface{}, query string, args ...interface{}) error {
	return db.sql.Get(dest, query, args...)
}

// write is not concurrency safe
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	db.write.Lock()
	defer db.write.Unlock()

	return db.sql.Exec(query, args...)
}
