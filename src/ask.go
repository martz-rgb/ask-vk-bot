package main

import (
	"github.com/leporo/sqlf"
)

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
	sqlf.SetDialect(sqlf.NoDialect)

	return &Ask{
		config: config,
		db:     db,
	}
}

func (a *Ask) Roles() ([]Role, error) {
	var roles []Role

	query := sqlf.From("roles").Bind(&Role{})
	err := a.db.Select(&roles, query.String())
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (a *Ask) RolesStartWith(prefix string) ([]Role, error) {
	var roles []Role

	query := sqlf.From("roles").Bind(&Role{}).Where("shown_name like ?", prefix+"%")
	err := a.db.Select(&roles, query.String(), query.Args()...)
	if err != nil {
		return nil, err
	}

	return roles, nil
}
