package ask

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

type ReservationStatus string

var ReservationStatuses = struct {
	UnderConsideration ReservationStatus
	InProgress         ReservationStatus
	Done               ReservationStatus
	Poll               ReservationStatus
}{
	UnderConsideration: "Under Consideration",
	InProgress:         "In Progress",
	Done:               "Done",
	Poll:               "Poll",
}

func (s ReservationStatus) Value() (driver.Value, error) {
	return string(s), nil
}

func (s *ReservationStatus) Scan(value interface{}) error {
	if value == nil {
		return errors.New("ReservationStatus is not nullable")
	}

	if str, err := driver.String.ConvertValue(value); err == nil {
		if v, ok := str.(string); ok {
			if v != string(ReservationStatuses.UnderConsideration) &&
				v != string(ReservationStatuses.InProgress) &&
				v != string(ReservationStatuses.Done) &&
				v != string(ReservationStatuses.Poll) {
				return errors.New("value is not valid ReservationStatus value")
			}

			*s = ReservationStatus(v)
			return nil
		}

	}
	return errors.New("failed to scan ReservationStatus")
}

type Urls []string

func (s Urls) Value() (driver.Value, error) {
	json, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	return string(json), nil
}

func (s *Urls) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}

	if str, err := driver.String.ConvertValue(value); err == nil {
		if v, ok := str.(string); ok {
			var urls []string
			err := json.Unmarshal([]byte(v), &urls)

			if err != nil {
				return errors.New("failed to unmarshal urls")
			}

			*s = urls
			return nil
		}

	}
	return errors.New("failed to scan Urls")

}

type Reservation struct {
	VkID         int               `db:"vk_id"`
	Introduction int               `db:"introduction"` // id of vk message contained information
	Deadline     sql.NullTime      `db:"deadline"`
	Status       ReservationStatus `db:"status"`
	Greeting     Urls              `db:"greeting"`
	Poll         sql.NullInt32     `db:"poll"`

	Role
}

func (a *Ask) AddReservation(vk_id int, role string, introduction int) error {
	query := sqlf.InsertInto("reservations").
		Set("vk_id", vk_id).
		Set("role", role).
		Set("introduction", introduction)

	if a.config.NoConfirmReservation {
		query.Set("is_confirmed", 1)
	}

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to add reservation",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}

func (a *Ask) ReservationByVkID(vk_id int) (*Reservation, error) {
	var reservations []Reservation

	query := sqlf.From("reservations_details").
		Bind(&Reservation{}).
		Where("vk_id = ?", vk_id).
		Limit(1)

	err := a.db.Select(&reservations, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get reservations by vk id",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	if len(reservations) == 0 {
		return nil, nil
	}

	// correct time
	reservations[0].Deadline.Time = reservations[0].Deadline.Time.Add(-a.timezone)
	return &reservations[0], nil
}

func (a *Ask) UnderConsiderationReservations() ([]Reservation, error) {
	var reservations []Reservation

	query := sqlf.From("reservations_details").
		Bind(&Reservation{}).
		Where("status = ?", ReservationStatuses.UnderConsideration)

	err := a.db.Select(&reservations, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get reservations under consideration",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return reservations, nil
}

func (a *Ask) InProgressReservations() ([]Reservation, error) {
	var reservations []Reservation

	query := sqlf.From("reservations_details").
		Bind(&Reservation{}).
		Where("status = ?", ReservationStatuses.InProgress)

	err := a.db.Select(&reservations, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get reservations in progress",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	for i := range reservations {
		reservations[i].Deadline.Time = reservations[i].Deadline.Time.Add(-a.timezone)
	}

	return reservations, nil
}

func (a *Ask) Reservations() ([]Reservation, error) {
	var reservations []Reservation

	query := sqlf.From("reservations_details").
		Bind(&Reservation{})

	err := a.db.Select(&reservations, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get reservations details",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	for i := range reservations {
		reservations[i].Deadline.Time = reservations[i].Deadline.Time.Add(-a.timezone)
	}

	return reservations, nil
}

func (a *Ask) CalculateReservationDeadline() time.Time {
	// got right date
	now := time.Now().
		UTC().
		Add(a.timezone)

	return time.Date(now.Year(),
		now.Month(),
		now.Day(),
		23,
		59,
		59,
		0,
		time.UTC).
		Add(a.config.ReservationDuration)
}

// func (a *Ask) ChangeReservationDeadline(vk_id int, deadline time.Time) error {
// 	query := sqlf.Update("reservations").
// 		Set("deadline", deadline).
// 		Where("vk_id = ?", vk_id)

// 	_, err := a.db.Exec(query.String(), query.Args()...)
// 	if err != nil {
// 		return zaperr.Wrap(err, "failed to change reservation deadline",
// 			zap.String("query", query.String()),
// 			zap.Any("args", query.Args()))
// 	}

// 	return nil
// }

// func (a *Ask) ChangeReservationDeadlineByRole(role string, deadline time.Time) error {
// 	query := sqlf.Update("reservations").
// 		Set("deadline", deadline).
// 		Where("role = ?", role)

// 	_, err := a.db.Exec(query.String(), query.Args()...)
// 	if err != nil {
// 		return zaperr.Wrap(err, "failed to change reservation deadline",
// 			zap.String("query", query.String()),
// 			zap.Any("args", query.Args()))
// 	}

// 	return nil
// }

func (a *Ask) ConfirmReservation(vk_id int) (time.Time, error) {
	deadline := a.CalculateReservationDeadline()

	confirm_query := sqlf.Update("reservations").
		Set("is_confirmed", 1).
		Where("vk_id = ?", vk_id)

	deadline_query := sqlf.With("updated_role",
		sqlf.From("reservations").
			Select("role").
			Where("vk_id = ?", vk_id)).
		Update("reservations").
		Set("deadline", deadline).
		Where("role IN updated_role")

	tx, err := a.db.NewTransaction()
	if err != nil {
		return time.Time{}, zaperr.Wrap(err, "failed to begin new transaction",
			zap.String("reason", "confirm reservation"))
	}

	_, err = tx.Exec(confirm_query.String(), confirm_query.Args()...)
	if err != nil {
		tx.Rollback()
		return time.Time{}, zaperr.Wrap(err, "failed to confirm reservation",
			zap.String("query", confirm_query.String()),
			zap.Any("args", confirm_query.Args()))
	}

	_, err = tx.Exec(deadline_query.String(), deadline_query.Args()...)
	if err != nil {
		tx.Rollback()
		return time.Time{}, zaperr.Wrap(err, "failed to update reservations' deadline",
			zap.String("query", deadline_query.String()),
			zap.Any("args", deadline_query.Args()))
	}

	err = tx.Commit()
	if err != nil {
		return time.Time{}, zaperr.Wrap(err, "failed to commit transaction",
			zap.String("reason", "confirm reservation"))
	}

	return deadline, nil
}

func (a *Ask) CompleteReservation(vk_id int, images Urls) error {
	query := sqlf.Update("reservations").
		Set("greeting", images).
		Where("vk_id = ?", vk_id)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to complete reservation",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}

func (a *Ask) DeleteReservation(vk_id int) error {
	query := sqlf.DeleteFrom("reservations").
		Where("vk_id = ?", vk_id)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to delete reservation by id",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}

func (a *Ask) DeleteReservationByDeadline(deadline time.Time) error {
	query := sqlf.DeleteFrom("reservations").
		Where("unixepoch(?) - unixepoch(deadline) > 0", deadline)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to delete reservation by deadline",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}

func (a *Ask) DeleteReservationByRole(role string) error {
	query := sqlf.DeleteFrom("reservations").
		Where("role = ?", role)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to delete reservation by role",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}
