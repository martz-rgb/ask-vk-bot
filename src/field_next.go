package main

import (
	"ask-bot/src/ask"
	"ask-bot/src/form"
	"ask-bot/src/vk"
	"fmt"

	"github.com/SevereCloud/vksdk/v2/api"
)

func ConfirmReservationField(value interface{}) (string, []*form.Field) {
	reservation, ok := value.(*ask.ReservationDetail)
	if !ok {
		return "", nil
	}

	text := fmt.Sprintf("Роль: %s\nСтраница: @id%d\n",
		reservation.AccusativeName,
		reservation.VkID)

	forward, err := vk.ForwardParam(
		reservation.VkID,
		[]int{reservation.Info})
	if err != nil {
		return "", nil
	}

	request := &vk.MessageParams{
		Text: text,
		Params: api.Params{
			"forward": forward,
		},
	}

	field := form.NewField(
		"action",
		request,
		ConfirmReservationOptions,
		nil,
		ConfirmReservationValidate,
		nil,
	)

	return "details", []*form.Field{field}
}
