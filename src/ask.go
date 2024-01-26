package main

import (
	"time"

	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

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

// roles
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

func (a *Ask) Role(name string) (Role, error) {
	var role Role

	query := sqlf.From("roles").Bind(&Role{}).Where("name == ?", name)
	err := a.db.Get(&role, query.String(), query.Args()...)
	if err != nil {
		return Role{}, zaperr.Wrap(err, "failed to get role",
			zap.String("name", name),
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return role, nil
}

// points
func (a *Ask) Points(id int) (int, error) {
	var points int

	// zero is default value, it is not a error if it is null
	query := sqlf.From("points").Select("COALESCE(SUM(diff), 0)").Where("person == ?", id)
	err := a.db.Get(&points, query.String(), query.Args()...)
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

// deadline
func (a *Ask) Deadline(member int) (time.Time, error) {
	var deadline int64

	// should be at least one record
	query := sqlf.From("deadline").Select("SUM(diff)").Where("member == ?", member)
	err := a.db.Get(&deadline, query.String(), query.Args()...)
	if err != nil {
		return time.Time{}, zaperr.Wrap(err, "failed to get deadline for memeber",
			zap.Int("member", member),
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return time.Unix(deadline, 0), nil
}

func (a *Ask) HistoryDeadline(member int) ([]Deadline, error) {
	var history []Deadline

	query := sqlf.From("deadline").Bind(&Deadline{}).Where("member == ?", member).OrderBy("timestamp DESC")
	err := a.db.Select(&history, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get history of deadline for member",
			zap.Int("member", member),
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return history, nil
}

// member
// TO-DO possible no member
func (a *Ask) MemberByRole(role string) (Member, error) {
	var member Member

	query := sqlf.From("members").Bind(&Member{}).Where("role == ?", role)
	err := a.db.Get(&member, query.String(), query.Args()...)
	if err != nil {
		return Member{}, zaperr.Wrap(err, "failed to get member by role",
			zap.String("role", role),
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return member, nil
}

func (a *Ask) MembersById(id int) ([]Member, error) {
	var members []Member

	query := sqlf.From("members").Bind(&Member{}).Where("person == ?", id)
	err := a.db.Select(&members, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get members by id",
			zap.Int("id", id),
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return members, nil
}
