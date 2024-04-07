package listener

import (
	"ask-bot/src/vk"
	"context"
	"fmt"
	"sync"
	"time"
)

func (l *Listener) RunDB(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(1 * time.Minute)

	for {
		select {
		case <-ticker.C:
			err := l.CheckReservationsDeadline()
			if err != nil {
				l.log.Errorw("failed to check reservation deadline",
					"error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (l *Listener) CheckReservationsDeadline() error {
	// check reservation deadlines
	reservations, err := l.c.Ask.InProgressReservationsDetails()
	if err != nil {
		return err
	}

	now := time.Now()
	notifications := []int{}
	for i := range reservations {
		if now.After(reservations[i].Deadline.Time) {
			notifications = append(notifications, i)
		}
	}

	err = l.c.Ask.DeleteReservationByDeadline(now)
	if err != nil {
		return err
	}

	for i := range notifications {
		message := &vk.MessageParams{
			Id: reservations[notifications[i]].VkID,
			Text: fmt.Sprintf("Ваша бронь на %s, к сожалению, закончилась! Вы можете забронировать роль снова или попробовать позже.",
				reservations[notifications[i]].AccusativeName),
		}
		l.c.NotifyUser <- message
	}

	return nil
}
