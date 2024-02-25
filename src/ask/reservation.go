package ask

import (
	"time"

	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

func (a *Ask) AddReservation(role string, vk_id int, info int) error {
	query := sqlf.InsertInto("reservations").
		Set("role", role).
		Set("vk_id", vk_id).
		Set("info", info)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to add reservation",
			zap.String("role", role),
			zap.Int("vk_id", vk_id),
			zap.String("query", query.String()),
			zap.Int("info", info),
			zap.Any("args", query.Args()))
	}

	return nil
}

func (a *Ask) ReservationsByVkID(vk_id int) ([]Reservation, error) {
	var reservations []Reservation

	query := sqlf.From("reservations").
		Bind(&Reservation{}).
		Where("vk_id", vk_id)
	err := a.db.Select(&reservations, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get reservations by vk id",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return reservations, nil
}

func (a *Ask) ReservationsDetailsByVkID(vk_id int) (*ReservationDetail, error) {
	var reservations_details []ReservationDetail

	query := sqlf.From("reservations_details").
		Bind(&ReservationDetail{}).
		Where("vk_id", vk_id).
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

func (a *Ask) UnderConsiderationReservationsDetails() ([]ReservationDetail, error) {
	var details []ReservationDetail

	query := sqlf.From("reservations_details").
		Bind(&ReservationDetail{}).
		Where("status == ?", ReservationStatuses.UnderConsideration)

	err := a.db.Select(&details, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get reservations details under consideration",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return details, nil
}

// func (a *Ask) ChangeReservationStatus(id int, status ReservationStatus) error {
// 	query := sqlf.Update("reservations").
// 		Set("status", status).
// 		Where("id == ?", id)

// 	_, err := a.db.Exec(query.String(), query.Args()...)
// 	if err != nil {
// 		return zaperr.Wrap(err, "failed to change reservation status",
// 			zap.String("query", query.String()),
// 			zap.Any("args", query.Args()))
// 	}

// 	return nil
// }

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
		Where("id", id)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to change reservation deadline",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}

func (a *Ask) ConfirmReservation(id int) (time.Time, error) {
	status := ReservationStatuses.InProgress
	deadline := a.CalculateReservationDeadline()

	query := sqlf.Update("reservations").
		Set("deadline", deadline).
		Set("status", status).
		Where("id", id)
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
		Set("status", ReservationStatuses.Done).
		Set("greeting", greeting).
		Where("id == ?", id)

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
		Where("id == ?", id)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to delete reservation",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}
