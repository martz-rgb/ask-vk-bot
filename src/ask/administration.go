package ask

import (
	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

// administration
func (a *Ask) IsAdmin(vk_id int) (bool, error) {
	var admin []Administration

	query := sqlf.From("administration").
		Bind(&Administration{}).
		Where("vk_id = ?", vk_id)

	err := a.db.Select(&admin, query.String(), query.Args()...)
	if err != nil {
		return false, zaperr.Wrap(err, "failed to get administration",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return len(admin) > 0, nil
}
