package ask

import (
	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

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
