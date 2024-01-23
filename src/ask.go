package main

import (
	"time"

	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
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

type Role struct {
	Name        string `db:"name"`
	Tag         string `db:"tag"`
	ShownName   string `db:"shown_name"`
	CaptionName string `db:"caption_name"`
	// Album       string `db:"album_link"`
	// Board       string `db:"board_link"`
}

type Points struct {
	Person    int       `db:"person"`
	Diff      int       `db:"diff"`
	Cause     string    `db:"cause"`
	Timestamp time.Time `db:"timestamp"`
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

// TO-DO should roles be sorted alphabetically or by groups
func (a *Ask) Roles() ([]Role, error) {
	var roles []Role

	query := sqlf.From("roles").Bind(&Role{})
	err := a.db.Select(&roles, query.String())
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get all roles",
			zap.String("query", query.String()))
	}

	return roles, nil
}

func (a *Ask) RolesStartWith(prefix string) ([]Role, error) {
	var roles []Role

	query := sqlf.From("roles").Bind(&Role{}).Where("shown_name like ?", prefix+"%")
	err := a.db.Select(&roles, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get roles starts with",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return roles, nil
}

func (a *Ask) Points(id int) (int, error) {
	var points int

	// zero is default value, it is not a error if it is null
	query := sqlf.From("points").Select("COALESCE(SUM(diff), 0)").Where("person == ?", id)
	err := a.db.QueryRow(&points, query.String(), query.Args()...)
	if err != nil {
		return -1, zaperr.Wrap(err, "failed to get points for user",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return points, nil
}

func (a *Ask) HistoryPoints(id int) ([]Points, error) {
	var history []Points

	query := sqlf.From("points").Bind(&Points{}).Where("person == ?", id).OrderBy("timestamp DESC")
	err := a.db.Select(&history, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get history of points for user",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return history, nil
}
