package ask

import (
	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

// member
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
