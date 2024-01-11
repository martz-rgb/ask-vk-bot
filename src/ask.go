package main

type AskConfig struct {
}

type Role struct {
	Name        string `db:"name"`
	Tag         string `db:"tag"`
	ShownName   string `db:"shown_name"`
	CaptionName string `db:"caption_name"`
}

type Ask struct {
	config *AskConfig
	db     *DB
}

func NewAsk(config *AskConfig, db *DB) *Ask {
	return &Ask{
		config: config,
		db:     db,
	}
}

func (a *Ask) Roles() ([]Role, error) {
	var roles []Role
	query := "select name, tag, shown_name, caption_name from roles"

	err := a.db.sql.Select(&roles, query)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (a *Ask) RolesStartWith(prefix string) ([]Role, error) {
	var roles []Role

	query := "select name, tag, shown_name, caption_name from roles where shown_name like ?"

	err := a.db.sql.Select(&roles, query, prefix+"%")
	if err != nil {
		return nil, err
	}

	return roles, nil
}
