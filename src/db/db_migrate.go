package db

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

type Object struct {
	Name string `db:"name"`
	Kind string `db:"type"`
	Sql  string `db:"sql"`
}

type Column struct {
	Name string `db:"name"`
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

	actual_tables := []Object{}
	desired_tables := []Object{}

	query := sqlf.From("sqlite_schema").
		Bind(&Object{}).
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

	new_tables, old_tables, modified_tables := getDifference(actual_tables, desired_tables)

	zap.S().Debugw("", "new", new_tables, "delete", old_tables, "modified", modified_tables)

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
	for _, table := range new_tables {
		zap.S().Infow("create new table",
			"name", table.Name,
			"statement", table.Sql)

		_, err := db.sql.Exec(table.Sql)
		if err != nil {
			transaction.Rollback()
			return zaperr.Wrap(err, "failed to create new table",
				zap.String("name", table.Name),
				zap.String("statement", table.Sql))
		}
	}

	// drop views to prevent errors for "not such table"
	// they will be restored after tables migration
	if len(old_tables) > 0 || len(modified_tables) > 0 {
		views := []Object{}

		query := sqlf.From("sqlite_master").
			Bind(&Object{}).
			Where("type = ?", "view").
			Where("sql IS NOT NULL").
			OrderBy("name")

		err := db.Select(&views, query.String(), query.Args()...)
		if err != nil {
			return zaperr.Wrap(err, "failed to get views",
				zap.String("statement", query.String()),
				zap.Any("args", query.Args()))
		}

		for _, view := range views {
			drop := fmt.Sprintf("DROP VIEW %s", view.Name)

			zap.S().Infow("delete view before modifying/deleting tables",
				"name", view.Name,
				"statement", drop)

			_, err := db.sql.Exec(drop)
			if err != nil {
				transaction.Rollback()
				return zaperr.Wrap(err, "failed to drop view before modifying/deleting tables",
					zap.String("name", view.Name),
					zap.String("statement", drop))
			}
		}
	}

	// delete tables
	for _, table := range old_tables {
		if allow_deletion {
			query := fmt.Sprintf("DROP TABLE %s", table.Name)

			zap.S().Infow("delete old table",
				"name", table.Name,
				"statement", query)

			_, err := db.sql.Exec(query)
			if err != nil {
				transaction.Rollback()
				return zaperr.Wrap(err, "failed to delete old table",
					zap.String("name", table.Name),
					zap.String("statement", query))
			}
		} else {
			zap.S().Infow("deletion is not allowed; skip removed table",
				"name", table.Name,
				"statement", table.Sql)
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

	// migrate indices, triggers and views
	err = db.migrateObjects("index", desired)
	if err != nil {
		transaction.Rollback()
		return err
	}

	err = db.migrateObjects("trigger", desired)
	if err != nil {
		transaction.Rollback()
		return err
	}

	err = db.migrateObjects("view", desired)
	if err != nil {
		transaction.Rollback()
		return err
	}

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
func (db *DB) migrateModified(desired *sqlx.DB, table Object, allow_deletion bool) error {
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

func (db *DB) migrateObjects(kind string, desired *sqlx.DB) error {
	actual_objects := []Object{}
	desired_objects := []Object{}

	query := sqlf.From("sqlite_master").
		Bind(&Object{}).
		Where("type = ?", kind).
		Where("sql IS NOT NULL").
		OrderBy("name")

	err := db.Select(&actual_objects, query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to get actual objects",
			zap.String("kind", kind),
			zap.String("statement", query.String()),
			zap.Any("args", query.Args()))
	}

	err = desired.Select(&desired_objects, query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to get desired objects",
			zap.String("kind", kind),
			zap.String("statement", query.String()),
			zap.Any("args", query.Args()))
	}

	new, old, modified := getDifference(actual_objects, desired_objects)

	for _, object := range new {
		if len(object.Sql) == 0 {
			continue
		}

		zap.S().Infow("create new object",
			"kind", kind,
			"name", object.Name,
			"statement", object.Sql)

		_, err = db.sql.Exec(object.Sql)
		if err != nil {
			return zaperr.Wrap(err, "failed to create new object",
				zap.String("kind", kind),
				zap.String("name", object.Name),
				zap.String("statement", object.Sql))
		}
	}

	for _, object := range old {
		query := fmt.Sprintf("DROP %s %s",
			strings.ToUpper(kind),
			object.Name)

		zap.S().Infow("delete old object",
			"kind", kind,
			"name", object.Name,
			"statement", object.Sql,
			"query", query)

		_, err = db.sql.Exec(query)
		if err != nil {
			return zaperr.Wrap(err, "failed to delete old object",
				zap.String("kind", kind),
				zap.String("name", object.Name),
				zap.String("statement", query))
		}
	}

	for _, object := range modified {
		if len(object.Sql) == 0 {
			continue
		}

		query := fmt.Sprintf("DROP %s %s",
			strings.ToUpper(kind),
			object.Name)

		zap.S().Infow("drop modified object",
			"kind", kind,
			"name", object.Name,
			"statement", object.Sql,
			"query", query)

		_, err = db.sql.Exec(query)
		if err != nil {
			return zaperr.Wrap(err, "failed to delete old modified object",
				zap.String("kind", kind),
				zap.String("index name", object.Name),
				zap.String("statement", query))
		}

		_, err = db.sql.Exec(object.Sql)

		zap.S().Infow("create new modified object",
			"kind", kind,
			"name", object.Name,
			"statement", object.Sql)

		if err != nil {
			return zaperr.Wrap(err, "failed to create new modified object",
				zap.String("kind", kind),
				zap.String("name", object.Name),
				zap.String("statement", object.Sql))
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

func getDifference(actual []Object, desired []Object) (new []Object, old []Object, modified []Object) {
	i, j := 0, 0
	for i < len(desired) && j < len(actual) {
		if desired[i].Name == actual[j].Name {
			if normalise(desired[i].Sql) != normalise(actual[j].Sql) {
				modified = append(modified, desired[i])
			}

			i, j = i+1, j+1
			continue
		}

		if desired[i].Name < actual[j].Name {
			new = append(new, desired[i])
			i++
		} else {
			old = append(old, actual[j])
			j++
		}
	}

	if i < len(desired) {
		new = append(new, desired[i:]...)
	}
	if j < len(actual) {
		old = append(old, actual[j:]...)
	}

	return
}
