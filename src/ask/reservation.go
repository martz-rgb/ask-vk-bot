package ask

import (
	"database/sql"
	"database/sql/driver"
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

type Reservation struct {
	Id          int            `db:"id"`
	Role        string         `db:"role"`
	VkID        int            `db:"vk_id"`
	Deadline    sql.NullTime   `db:"deadline"`
	IsConfirmed int            `db:"is_confirmed"`
	Info        int            `db:"info"` // id of vk message contained information
	Greeting    sql.NullString `db:"greeting"`
	Timestamp   time.Time      `db:"timestamp"`
}

type ReservationDetails struct {
	Reservation
	Role

	Status ReservationStatus `db:"status"`
	Post   sql.NullInt32     `db:"post"`
}

func (a *Ask) AddReservation(role string, vk_id int, info int) error {
	query := sqlf.InsertInto("reservations").
		Set("role", role).
		Set("vk_id", vk_id).
		Set("info", info)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to add reservation",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}

func (a *Ask) ReservationDetailsByVkID(vk_id int) (*ReservationDetails, error) {
	var reservations_details []ReservationDetails

	query := sqlf.From("reservations_details").
		Bind(&ReservationDetails{}).
		Where("vk_id = ?", vk_id).
		Limit(1)

	err := a.db.Select(&reservations_details, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get reservations details by vk id",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	if len(reservations_details) == 0 {
		return nil, nil
	}

	return &reservations_details[0], nil
}

func (a *Ask) UnderConsiderationReservationsDetails() ([]ReservationDetails, error) {
	var details []ReservationDetails

	query := sqlf.From("reservations_details").
		Bind(&ReservationDetails{}).
		Where("status = ?", ReservationStatuses.UnderConsideration)

	err := a.db.Select(&details, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get reservations details under consideration",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return details, nil
}

func (a *Ask) ReservationsDetails() ([]ReservationDetails, error) {
	var details []ReservationDetails

	query := sqlf.From("reservations_details").
		Bind(&ReservationDetails{})

	err := a.db.Select(&details, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get reservations details",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return details, nil
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

func (a *Ask) ChangeReservationDeadline(id int, deadline time.Time) error {
	query := sqlf.Update("reservations").
		Set("deadline", deadline).
		Where("id = ?", id)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to change reservation deadline",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}

// TO-DO update also other deadlines on role
// transaction logic for db...
func (a *Ask) ConfirmReservation(id int) (time.Time, error) {
	deadline := a.CalculateReservationDeadline()

	query := sqlf.Update("reservations").
		Set("deadline", deadline).
		Set("is_confirmed", 1).
		Where("id = ?", id)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return time.Time{}, zaperr.Wrap(err, "failed to confirm reservation",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return deadline, nil
}

func (a *Ask) CompleteReservation(id int, greeting string) error {
	query := sqlf.Update("reservations").
		Set("greeting", greeting).
		Where("id = ?", id)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to complete reservation",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}

func (a *Ask) DeleteReservation(id int) error {
	query := sqlf.DeleteFrom("reservations").
		Where("id = ?", id)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to delete reservation",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}
