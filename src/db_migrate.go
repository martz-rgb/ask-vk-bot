package main

// main source: https://david.rothlis.net/declarative-schema-migration-for-sqlite/

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hori-ryota/zaperr"
	"github.com/jmoiron/sqlx"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

type Table struct {
	Name string `db:"name"`
	Sql  string `db:"sql"`
}

type Column struct {
	Name string `db:"name"`
}

type Index struct {
	Name string `db:"name"`
	Sql  string `db:"sql"`
}

func (db *DB) Migrate(schema string, allow_deletion bool) error {
	desired, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		return zaperr.Wrap(err, "failed to open in-memory database")
	}

	_, err = desired.Exec(string(schema))
	if err != nil {
		return zaperr.Wrap(err, "failed to execute schema on pristine database",
			zap.String("schema", schema))
	}

	actual_tables := []Table{}
	desired_tables := []Table{}

	query := sqlf.From("sqlite_schema").
		Bind(&Table{}).
		Where("type = 'table'").
		Where("name != 'sqlite_sequence'").
		OrderBy("name")

	err = db.Select(&actual_tables, query.String())
	if err != nil {
		return zaperr.Wrap(err, "failed to get actual tables",
			zap.String("statement", query.String()),
			zap.Any("args", query.Args()))
	}

	err = desired.Select(&desired_tables, query.String())
	if err != nil {
		return zaperr.Wrap(err, "failed to get desired tables",
			zap.String("statement", query.String()),
			zap.Any("args", query.Args()))
	}

	new_tables := []Table{}
	delete_tables := []Table{}
	modified_tables := []Table{}

	i, j := 0, 0
	for i < len(desired_tables) && j < len(actual_tables) {
		if desired_tables[i].Name == actual_tables[j].Name {
			if normalise(desired_tables[i].Sql) != normalise(actual_tables[j].Sql) {
				modified_tables = append(modified_tables, desired_tables[i])
			}

			i, j = i+1, j+1
			continue
		}

		if desired_tables[i].Name < actual_tables[j].Name {
			new_tables = append(new_tables, desired_tables[i])
			i++
		} else {
			delete_tables = append(delete_tables, actual_tables[j])
			j++
		}
	}

	if i < len(desired_tables) {
		new_tables = append(new_tables, desired_tables[i:]...)
	}
	if j < len(actual_tables) {
		delete_tables = append(delete_tables, actual_tables[j:]...)
	}

	zap.S().Debugw("", "new", new_tables, "delete", delete_tables, "modified", modified_tables)

	// start process
	_, err = db.sql.Exec("PRAGMA foreign_keys=off")
	if err != nil {
		return zaperr.Wrap(err, "failed to disable foreign_keys")
	}

	transaction, err := db.sql.Begin()
	if err != nil {
		return zaperr.Wrap(err, "failed to start transaction")
	}

	// create new tables
	for i := range new_tables {
		zap.S().Infow("create new table",
			"name", new_tables[i].Name,
			"statement", new_tables[i].Sql)

		_, err := db.sql.Exec(new_tables[i].Sql)
		if err != nil {
			transaction.Rollback()
			return zaperr.Wrap(err, "failed to create new table",
				zap.String("name", new_tables[i].Name),
				zap.String("statement", new_tables[i].Sql))
		}
	}

	// delete tables
	for i := range delete_tables {
		if allow_deletion {
			query := fmt.Sprintf("DROP TABLE %s", delete_tables[i].Name)

			zap.S().Infow("delete old table",
				"name", delete_tables[i].Name,
				"statement", query)

			_, err := db.sql.Exec(query)
			if err != nil {
				transaction.Rollback()
				return zaperr.Wrap(err, "failed to delete old table",
					zap.String("name", delete_tables[i].Name),
					zap.String("statement", query))
			}
		} else {
			zap.S().Infow("deletion is not allowed; skip removed table",
				"name", delete_tables[i].Name,
				"statement", delete_tables[i].Sql)
		}
	}

	// modify tables
	for i := range modified_tables {
		err := db.migrateModified(desired, modified_tables[i], allow_deletion)
		if err != nil {
			transaction.Rollback()
			return err
		}
	}

	err = db.migrateIndices(desired, allow_deletion)
	if err != nil {
		transaction.Rollback()
		return err
	}

	// add views and trigger probably

	err = transaction.Commit()
	if err != nil {
		return zaperr.Wrap(err, "failed to commit transaction")
	}

	_, err = db.sql.Exec("PRAGMA foreign_keys=on")
	if err != nil {
		return zaperr.Wrap(err, "failed to enable back foreign_keys")
	}

	return nil
}

// procedure: https://www.sqlite.org/lang_altertable.html#otheralter
func (db *DB) migrateModified(desired *sqlx.DB, table Table, allow_deletion bool) error {
	// check if no removing is meant
	actual_cols := []Column{}
	desired_cols := []Column{}

	query := sqlf.From(fmt.Sprintf("pragma_table_info('%s')", table.Name)).
		Bind(&Column{}).
		OrderBy("name")

	err := db.Select(&actual_cols, query.String())
	if err != nil {
		return zaperr.Wrap(err, "failed to get actual table info",
			zap.String("name", table.Name),
			zap.String("statement", query.String()))
	}
	err = desired.Select(&desired_cols, query.String())
	if err != nil {
		return zaperr.Wrap(err, "failed to get desired table info",
			zap.String("name", table.Name),
			zap.String("statement", query.String()))
	}

	left_cols := []Column{}
	i, j := 0, 0
	for i < len(desired_cols) && j < len(actual_cols) {
		if desired_cols[i].Name < actual_cols[j].Name {
			i++
			continue
		}
		if desired_cols[i].Name > actual_cols[j].Name {
			j++
			continue
		}

		left_cols = append(left_cols, actual_cols[j])
		i, j = i+1, j+1
	}

	if len(left_cols) < len(actual_cols) && !allow_deletion {
		zap.S().Infow("deletion is not allowed; skip modified table with removed cols",
			"name", table.Name,
			"statement", table.Sql,
			"actual columns", actual_cols,
			"desired columns", desired_cols)
		return nil
	}

	// create new table
	new_name := fmt.Sprintf("%s_migration", table.Name)
	create_rename := regexp.MustCompile(fmt.Sprintf(`\b%s\b`, table.Name)).
		ReplaceAllString(table.Sql, new_name)

	zap.S().Infow("create new table for migration",
		"name", table.Name,
		"new_name", new_name,
		"statement", create_rename)

	_, err = db.sql.Exec(create_rename)
	if err != nil {
		return zaperr.Wrap(err, "failed to create renamed table",
			zap.String("name", table.Name),
			zap.String("new_name", new_name),
			zap.String("statement", create_rename))
	}

	// migrate data
	cols := []string{}
	for _, col := range left_cols {
		cols = append(cols, col.Name)
	}
	list := strings.Join(cols, ",")

	insert := fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM %s", new_name, list, list, table.Name)

	zap.S().Infow("migrate data from old table to new",
		"name", table.Name,
		"new_name", new_name,
		"statement", insert)

	_, err = db.sql.Exec(insert)
	if err != nil {
		return zaperr.Wrap(err, "failed to migrate data from old table to new one",
			zap.String("name", table.Name),
			zap.String("new_name", new_name),
			zap.String("statement", insert),
			zap.Any("actual columns", actual_cols),
			zap.Any("desired columns", desired_cols))
	}

	// drop old table
	drop := fmt.Sprintf("DROP TABLE %s", table.Name)

	zap.S().Infow("drop old modified table",
		"name", table.Name,
		"statement", drop)

	_, err = db.sql.Exec(drop)
	if err != nil {
		return zaperr.Wrap(err, "failed to drop old table",
			zap.String("name", table.Name),
			zap.String("statement", drop))
	}

	// change name
	alter := fmt.Sprintf("ALTER TABLE %s RENAME TO %s", new_name, table.Name)

	zap.S().Infow("rename new modified table",
		"name", table.Name,
		"old_name", new_name,
		"statement", alter)

	_, err = db.sql.Exec(alter)
	if err != nil {
		return zaperr.Wrap(err, "failed to rename table",
			zap.String("name", table.Name),
			zap.String("old_name", new_name),
			zap.String("statement", alter))
	}

	return nil
}

func (db *DB) migrateIndices(desired *sqlx.DB, allow_deletion bool) error {
	// migrate indices
	actual_indices := []Index{}
	desired_indices := []Index{}

	// pnly user-defined indices
	query := sqlf.From("sqlite_master").
		Bind(&Index{}).
		Where("type = 'index'").
		Where("sql IS NOT NULL").
		OrderBy("name")

	err := db.Select(&actual_indices, query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to get actual indices",
			zap.String("statement", query.String()),
			zap.Any("args", query.Args()))
	}

	err = desired.Select(&desired_indices, query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to get desired indices",
			zap.String("statement", query.String()),
			zap.Any("args", query.Args()))
	}

	new_indices := []Index{}
	delete_indices := []Index{}
	modified_indices := []Index{}

	i, j := 0, 0
	for i < len(desired_indices) && j < len(actual_indices) {
		if desired_indices[i].Name == actual_indices[j].Name {
			if normalise(desired_indices[i].Sql) != normalise(actual_indices[j].Sql) {
				modified_indices = append(modified_indices, desired_indices[i])
			}

			i, j = i+1, j+1
			continue
		}

		if desired_indices[i].Name < actual_indices[j].Name {
			new_indices = append(new_indices, desired_indices[i])
			i++
		} else {
			delete_indices = append(delete_indices, actual_indices[j])
			j++
		}
	}

	if i < len(desired_indices) {
		new_indices = append(new_indices, desired_indices[i:]...)
	}
	if j < len(actual_indices) {
		delete_indices = append(delete_indices, actual_indices[j:]...)
	}

	for _, index := range new_indices {
		if len(index.Sql) == 0 {
			continue
		}

		zap.S().Infow("create new index",
			"name", index.Name,
			"statement", index.Sql)

		_, err = db.sql.Exec(index.Sql)
		if err != nil {
			return zaperr.Wrap(err, "failed to create new index",
				zap.String("index name", index.Name),
				zap.String("statement", index.Sql))
		}
	}

	for _, index := range delete_indices {
		query := fmt.Sprintf("DROP INDEX %s", index.Name)

		zap.S().Infow("delete old index",
			"name", index.Name,
			"statement", index.Sql,
			"query", query)

		_, err = db.sql.Exec(query)
		if err != nil {
			return zaperr.Wrap(err, "failed to delete old index",
				zap.String("index name", index.Name),
				zap.String("statement", query))
		}
	}

	for _, index := range modified_indices {
		if len(index.Sql) == 0 {
			continue
		}

		query := fmt.Sprintf("DROP INDEX %s", index.Name)

		zap.S().Infow("drop modified index",
			"name", index.Name,
			"statement", index.Sql,
			"query", query)

		_, err = db.sql.Exec(query)
		if err != nil {
			return zaperr.Wrap(err, "failed to delete old modified index",
				zap.String("index name", index.Name),
				zap.String("statement", query))
		}

		_, err = db.sql.Exec(index.Sql)

		zap.S().Infow("create new modified index",
			"name", index.Name,
			"statement", index.Sql)

		if err != nil {
			return zaperr.Wrap(err, "failed to create new modified index",
				zap.String("index name", index.Name),
				zap.String("statement", index.Sql))
		}
	}
	return nil
}

func normalise(query string) string {
	// to lowercase
	query = strings.ToLower(query)

	// comments
	query = regexp.MustCompile(`--[^\n]*\n`).ReplaceAllString(query, ``)

	// whitespace
	query = regexp.MustCompile(`\s+`).ReplaceAllString(query, ` `)
	query = regexp.MustCompile(` *([(),]) *`).ReplaceAllString(query, `$1`)

	//  unnecessary quotes
	query = regexp.MustCompile(`\'(\w+)\'`).ReplaceAllString(query, `$1`)
	query = regexp.MustCompile(`\"(\w+)\"`).ReplaceAllString(query, `$1`)

	return strings.TrimSpace(query)
}
