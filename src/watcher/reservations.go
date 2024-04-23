package watcher

import (
	"ask-bot/src/vk"
	"fmt"
	"time"
)

func (c *Controls) CheckReservationsDeadline() error {

	reservations, err := c.Ask.InProgressReservations()
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

	err = c.Ask.DeleteReservationByDeadline(now)
	if err != nil {
		return err
	}

	for i := range notifications {
		message := &vk.MessageParams{
			Id: reservations[notifications[i]].VkID,
			Text: fmt.Sprintf("Ваша бронь на %s, к сожалению, закончилась! Вы можете забронировать роль снова или попробовать позже.",
				reservations[notifications[i]].AccusativeName),
		}
		c.NotifyUser <- message
	}

	return nil

}
