package main

import (
	"time"

	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

type Ask struct {
	config *AskConfig
	db     *DB

	timezone time.Duration
}

func NewAsk(config *AskConfig, db *DB) *Ask {
	sqlf.SetDialect(sqlf.NoDialect)

	return &Ask{
		config:   config,
		db:       db,
		timezone: time.Duration(config.Timezone) * time.Hour,
	}
}

// administration
func (a *Ask) IsAdmin(vk_id int) (bool, error) {
	var admin []Administration

	query := sqlf.From("administration").
		Bind(&Administration{}).
		Where("vk_id == ?", vk_id)

	err := a.db.Select(&admin, query.String(), query.Args()...)
	if err != nil {
		return false, zaperr.Wrap(err, "failed to get administration",
			zap.String("query", query.String()))
	}

	return len(admin) > 0, nil
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

func (a *Ask) AvailableRoles() ([]Role, error) {
	var roles []Role

	query := sqlf.From("roles").
		Bind(&Role{}).
		With("busy_roles",
			sqlf.From("members").
				Select("role")).
		With("reserved_roles",
			sqlf.From("reservations").
				Select("role").
				Where("status == ?", ReservationStatuses.Done)).
		Where("name NOT IN busy_roles").
		Where("name NOT IN reserved_roles")

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

	query := sqlf.From("roles").Bind(&Role{}).Where("shown_name like ?", prefix+"%")
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
	query := sqlf.From("roles").
		Bind(&Role{}).
		With("busy_roles",
			sqlf.From("members").
				Select("role")).
		With("reserved_roles",
			sqlf.From("reservations").
				Select("role").
				Where("status == ?", ReservationStatuses.Done)).
		Where("name NOT IN busy_roles").
		Where("name NOT IN reserved_roles").
		Where("shown_name like ?", prefix+"%")

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
func (a *Ask) Points(vk_id int) (int, error) {
	var points int

	// zero is default value, it is not a error if it is null
	query := sqlf.From("points").Select("COALESCE(SUM(diff), 0)").Where("vk_id == ?", vk_id)
	err := a.db.Get(&points, query.String(), query.Args()...)
	if err != nil {
		return -1, zaperr.Wrap(err, "failed to get points for user",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return points, nil
}

func (a *Ask) HistoryPoints(vk_id int) ([]Points, error) {
	var history []Points

	query := sqlf.From("points").Bind(&Points{}).Where("vk_id == ?", vk_id).OrderBy("timestamp DESC")
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

	return time.Unix(deadline, 0).Add(a.timezone), nil
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

// TO-DO maybe another way to insert
func (a *Ask) ChangeDeadline(member int, diff time.Duration, kind DeadlineCause, cause string) error {
	query := sqlf.InsertInto("deadline").
		Set("member", member).
		Set("diff", diff.Seconds()).
		Set("kind", kind).
		Set("cause", cause)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to insert deadline event",
			zap.Int("member", member),
			zap.Duration("diff", diff),
			zap.Any("kind", kind),
			zap.String("cause", cause),
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
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

func (a *Ask) MembersByVkID(vk_id int) ([]Member, error) {
	var members []Member

	query := sqlf.From("members").Bind(&Member{}).Where("vk_id == ?", vk_id)
	err := a.db.Select(&members, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get members by vk_id",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return members, nil
}

func (a *Ask) AddMember(vk_id int, role string) error {
	query := sqlf.InsertInto("members").
		Set("vk_id", vk_id).
		Set("role", role)

	result, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to add member",
			zap.Int("vk_id", vk_id),
			zap.String("role", role),
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	member, err := result.LastInsertId()
	if err != nil {
		return zaperr.Wrap(err, "failed to get last inserted id",
			zap.Int("vk_id", vk_id),
			zap.String("role", role),
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	// init deadline
	return a.ChangeDeadline(int(member),
		a.config.Deadline,
		DeadlineCauses.Init,
		"init deadline")
}

// reservations
func (a *Ask) AddReservation(role string, vk_id int, info int) (time.Time, error) {
	// got right date
	now := time.Now().
		UTC().
		Add(a.timezone)

	deadline := time.Date(now.Year(),
		now.Month(),
		now.Day(),
		23,
		59,
		59,
		0,
		time.UTC).
		Add(a.config.ReservationDuration)

	query := sqlf.InsertInto("reservations").
		Set("role", role).
		Set("vk_id", vk_id).
		Set("deadline", deadline).
		Set("info", info)
	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return time.Time{}, zaperr.Wrap(err, "failed to add reservation",
			zap.String("role", role),
			zap.Int("vk_id", vk_id),
			zap.Time("deadline", deadline),
			zap.String("query", query.String()),
			zap.Int("info", info),
			zap.Any("args", query.Args()))
	}

	return deadline, nil
}

func (a *Ask) UnderConsiderationReservations() ([]Reservation, error) {
	var reservations []Reservation

	query := sqlf.From("reservations").
		Bind(&Reservation{}).
		Where("status == ?", ReservationStatuses.UnderConsideration)

	err := a.db.Select(reservations, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get reservations under consideration",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return reservations, nil
}

func (a *Ask) ChangeReservationStatus(id int, status ReservationStatus) error {
	query := sqlf.Update("reservations").
		Set("status", status).
		Where("id == ?", id)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to change reservation status",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}

func (a *Ask) DeleteReservation(id int) error {
	query := sqlf.DeleteFrom("reservation").
		Where("id == ?", id)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to delete reservation",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}
