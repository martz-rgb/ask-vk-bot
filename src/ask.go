package main

import (
	"time"

	"github.com/leporo/sqlf"
)

type Administration struct {
	VkID int `db:"vk_id"`
}

// TO-DO: time.Time is dangerous but i want to try
type Person struct {
	VkID     int       `db:"vk_id"`
	Name     string    `db:"name"`
	Gallery  string    `db:"gallery"`
	Birthday time.Time `db:"birthday"`
}

type AskConfig struct {
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

type Role struct {
	Name        string `db:"name"`
	Tag         string `db:"tag"`
	ShownName   string `db:"shown_name"`
	CaptionName string `db:"caption_name"`
	// Album       string `db:"album_link"`
	// Board       string `db:"board_link"`
}

// TO-DO should roles be sorted alphabetically or by groups
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

type Points struct {
	Id        int       `db:"id"`
	Person    int       `db:"person"`
	Diff      int       `db:"diff"`
	Cause     string    `db:"cause"`
	Timestamp time.Time `db:"timestamp"`
}

func (a *Ask) Points(id int) (int, error) {
	var points int

	// zero is default value, it is not a error if it is null
	query := sqlf.From("points").Select("COALESCE(SUM(diff), 0)").Where("id == ?", id)
	err := a.db.QueryRow(&points, query.String(), query.Args()...)
	if err != nil {
		return 0, err
	}

	return points, nil
}
