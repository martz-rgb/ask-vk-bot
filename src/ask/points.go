package ask

import (
	"time"

	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

type Points struct {
	VkID      int       `db:"vk_id"`
	Diff      int       `db:"diff"`
	Cause     string    `db:"cause"`
	Timestamp time.Time `db:"timestamp"`
}

func (a *Ask) PointsByVkID(vk_id int) (int, error) {
	var points int

	// zero is default value, it is not a error if it is null
	query := sqlf.From("points").
		Select("COALESCE(SUM(diff), 0)").
		Where("vk_id = ?", vk_id)

	err := a.db.Get(&points, query.String(), query.Args()...)
	if err != nil {
		return -1, zaperr.Wrap(err, "failed to get points by vk id",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return points, nil
}

func (a *Ask) HistoryPointsByVkID(vk_id int) ([]Points, error) {
	var history []Points

	query := sqlf.From("points").
		Bind(&Points{}).
		Where("vk_id = ?", vk_id).
		OrderBy("timestamp DESC")

	err := a.db.Select(&history, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get history of points by vk id",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return history, nil
}
