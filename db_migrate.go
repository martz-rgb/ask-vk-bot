package main

// main source: https://david.rothlis.net/declarative-schema-migration-for-sqlite/

import (
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type DbTable struct {
	Name string `db:"name"`
	Sql  string `db:"sql"`
}

func (db *Db) Migrate(schema string) error {
	pristine, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		return err
	}

	_, err = pristine.Exec(string(schema))
	if err != nil {
		return err
	}

	query := `select name, sql from sqlite_schema where type = 'table' and name != 'sqlite_sequence' order by name`

	pristine_tables := []DbTable{}
	err = pristine.Select(&pristine_tables, query)
	if err != nil {
		return err
	}

	actual_tables := []DbTable{}
	err = db.sql.Select(&actual_tables, query)
	if err != nil {
		return err
	}

	new_tables := []DbTable{}
	delete_tables := []DbTable{}
	modified_tables := []DbTable{}

	i, j := 0, 0
	for i < len(pristine_tables) && j < len(actual_tables) {
		if pristine_tables[i].Name == actual_tables[j].Name {
			if normalise(pristine_tables[i].Sql) != normalise(actual_tables[j].Sql) {
				modified_tables = append(modified_tables, pristine_tables[i])
			}

			i, j = i+1, j+1
			continue
		}

		if pristine_tables[i].Name < actual_tables[j].Name {
			new_tables = append(new_tables, pristine_tables[i])
			i++
		} else {
			delete_tables = append(delete_tables, actual_tables[j])
			j++
		}
	}

	if i < len(pristine_tables) {
		new_tables = append(new_tables, pristine_tables[i:]...)
	}
	if j < len(actual_tables) {
		delete_tables = append(delete_tables, actual_tables[j:]...)
	}

	zap.S().Debugw("", "new", new_tables, "delete", delete_tables, "modified", modified_tables)

	return nil
}

// TO-DO check if it is okay...
func normalise(query string) string {
	// comments
	query = regexp.MustCompile(`--[^\n]*\n`).ReplaceAllString(query, "")

	// whitespace
	query = regexp.MustCompile(`\s+`).ReplaceAllString(query, " ")
	query = regexp.MustCompile(` *([(),]) *`).ReplaceAllString(query, "$1")

	//  unnecessary quotes
	query = regexp.MustCompile(`\'(\w+)\'`).ReplaceAllString(query, "$1")
	query = regexp.MustCompile(`\"(\w+)\"`).ReplaceAllString(query, "$1")

	return strings.TrimSpace(query)
}
