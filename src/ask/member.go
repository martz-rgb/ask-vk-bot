package ask

import (
	"database/sql/driver"
	"errors"

	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

type MemberStatus string

var MemberStatuses = struct {
	Active MemberStatus
	Freeze MemberStatus
}{
	Active: "Active",
	Freeze: "Freeze",
}

func (s MemberStatus) Value() (driver.Value, error) {
	return string(s), nil
}

func (s *MemberStatus) Scan(value interface{}) error {
	if value == nil {
		return errors.New("MemberStatus is not nullable")
	}
	if str, err := driver.String.ConvertValue(value); err == nil {
		if v, ok := str.(string); ok {
			// check if is valid
			if v != string(MemberStatuses.Active) && v != string(MemberStatuses.Freeze) {
				return errors.New("value is not valid MemberStatus value")
			}
			*s = MemberStatus(v)
			return nil
		}
	}
	return errors.New("failed to scan MemberStatus")
}

type Member struct {
	Id       int          `db:"id"`
	VkID     int          `db:"vk_id"`
	Role     string       `db:"role"`
	Status   MemberStatus `db:"status"`
	Timezone int          `db:"timezone"`
}

// TO-DO possible no member
func (a *Ask) MemberByRole(role string) (Member, error) {
	var member Member

	query := sqlf.From("members").
		Bind(&Member{}).
		Where("role = ?", role)

	err := a.db.Get(&member, query.String(), query.Args()...)
	if err != nil {
		return Member{}, zaperr.Wrap(err, "failed to get member by role",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return member, nil
}

func (a *Ask) MembersByVkID(vk_id int) ([]Member, error) {
	var members []Member

	query := sqlf.From("members").
		Bind(&Member{}).
		Where("vk_id = ?", vk_id)

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
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	member, err := result.LastInsertId()
	if err != nil {
		return zaperr.Wrap(err, "failed to get last inserted id",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	// init deadline
	return a.ChangeDeadline(int(member),
		a.config.Deadline,
		DeadlineCauses.Init,
		"init deadline")
}
