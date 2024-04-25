package ask

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

type Role struct {
	Name           string         `db:"name"`
	Hashtag        string         `db:"hashtag"`
	ShownName      string         `db:"shown_name"`
	AccusativeName string         `db:"accusative_name"`
	CaptionName    string         `db:"caption_name"`
	Group          sql.NullString `db:"[group]"`
	Order          sql.NullInt32  `db:"[order]"`
	Album          sql.NullInt32  `db:"album"`
	Board          sql.NullInt32  `db:"board"`
}

// TO-DO should roles be sorted alphabetically or by groups
func (a *Ask) Roles() ([]Role, error) {
	var roles []Role

	query := sqlf.From("roles").
		Bind(&Role{})

	err := a.db.Select(&roles, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get roles",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return roles, nil
}

func (a *Ask) AvailableRoles() ([]Role, error) {
	var roles []Role

	query := sqlf.From("available_roles").
		Bind(&Role{})

	err := a.db.Select(&roles, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get available roles",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return roles, nil
}

func (a *Ask) RolesStartWith(prefix string) ([]Role, error) {
	var roles []Role

	query := sqlf.From("roles").
		Bind(&Role{}).
		Where("shown_name like ?", prefix+"%")

	err := a.db.Select(&roles, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get roles starts with",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return roles, nil
}

func (a *Ask) AvailableRolesStartWith(prefix string) ([]Role, error) {
	var roles []Role
	query := sqlf.From("available_roles").
		Bind(&Role{}).
		Where("shown_name like ?", prefix+"%")

	err := a.db.Select(&roles, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get available roles starts with",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return roles, nil
}

func (a *Ask) Role(name string) (Role, error) {
	var role Role

	query := sqlf.From("roles").
		Bind(&Role{}).
		Where("name = ?", name)

	err := a.db.Get(&role, query.String(), query.Args()...)
	if err != nil {
		return Role{}, zaperr.Wrap(err, "failed to get role",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return role, nil
}

// roles order by hashtags
func (a *Ask) RolesDictionary() ([]Role, error) {
	var roles []Role

	query := sqlf.From("roles").
		Bind(&Role{}).
		OrderBy("hashtag")

	err := a.db.Select(&roles, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get roles dictionary",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return roles, nil
}

// unused!
type MatchedHashtag struct {
	Hashtag string         `db:"hashtag"`
	Role    sql.NullString `db:"role"`
}

func (a *Ask) MatchHashtags(hashtags []string) ([]MatchedHashtag, error) {
	var matched []MatchedHashtag

	values := make([]string, len(hashtags))
	args := make([]interface{}, len(hashtags))

	for i := range hashtags {
		values[i] = "(?)"
		args[i] = hashtags[i]
	}

	subquery := sqlf.New(
		fmt.Sprintf("VALUES %s", strings.Join(values, ",")),
		args...,
	)

	query := sqlf.With("hashtags(value)", subquery).
		From("hashtags").
		LeftJoin("roles", "lower(hashtags.value) = lower(roles.hashtag)").
		Select("hashtags.value as hashtag").
		Select("roles.name as role").
		OrderBy("hashtag")

	err := a.db.Select(&matched, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to match hashtags",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return matched, nil
}

func (a *Ask) ChangeAlbums(albums map[string]int) error {
	keys := make([]interface{}, len(albums))
	values := make([]string, len(albums))

	index := 0
	for key, value := range albums {
		keys[index] = key
		values[index] = fmt.Sprintf("WHEN '%s' THEN %d", key, value)

		index++
	}

	params := strings.Repeat("?,", len(keys))

	query := sqlf.Update("roles").
		Clause(fmt.Sprintf("SET album = CASE name %s END", strings.Join(values, " "))).
		Where(fmt.Sprintf("name IN (%s)", params[:len(params)-1]), keys...)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to change albums",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}

func (a *Ask) ChangeBoards(boards map[string]int) error {
	keys := make([]interface{}, len(boards))
	values := make([]string, len(boards))

	index := 0
	for key, value := range boards {
		keys[index] = key
		values[index] = fmt.Sprintf("WHEN '%s' THEN %d", key, value)

		index++
	}

	params := strings.Repeat("?,", len(keys))

	query := sqlf.Update("roles").
		Clause(fmt.Sprintf("SET board = CASE name %s END", strings.Join(values, " "))).
		Where(fmt.Sprintf("name IN (%s)", params[:len(params)-1]), keys...)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to change boards",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}
