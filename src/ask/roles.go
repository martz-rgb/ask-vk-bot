package ask

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

// roles
// TO-DO should roles be sorted alphabetically or by groups
func (a *Ask) Roles() ([]Role, error) {
	var roles []Role

	query := sqlf.From("roles").
		Bind(&Role{})

	err := a.db.Select(&roles, query.String())
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
