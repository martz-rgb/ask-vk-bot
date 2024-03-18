package fields

import (
	"ask-bot/src/ask"
	"ask-bot/src/dict"
	"ask-bot/src/vk"
	"fmt"
)

func ConfirmReservationDeletion(values dict.Dictionary) (*vk.MessageParams, error) {
	reservation, err := dict.ExtractValue[*ask.ReservationDetails](values, "reservation")
	if err != nil {
		return nil, err
	}

	return &vk.MessageParams{
		Text: fmt.Sprintf("Вы уверены, что хотите удалить бронь на %s от @id%d?",
			reservation.AccusativeName,
			reservation.VkID),
	}, nil
}
