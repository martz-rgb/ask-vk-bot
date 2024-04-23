package fields

import (
	"ask-bot/src/ask"
	"ask-bot/src/chatbot/states/validate"
	"ask-bot/src/form"
	"ask-bot/src/vk"
	"fmt"

	"github.com/SevereCloud/vksdk/v2/api"
)

// (admin) confirm reservation
var ConfirmReservationOptions = []form.Option{
	{
		ID:    "confirm",
		Label: "Потвердить",
		Value: true,
	},
	{
		ID:    "delete",
		Label: "Удалить",
		Value: false,
	},
}

func ConfirmReservationField(value interface{}) (string, []*form.Field) {
	reservation, ok := value.(*ask.Reservation)
	if !ok {
		return "", nil
	}

	text := fmt.Sprintf("Роль: %s\nСтраница: @id%d\n",
		reservation.AccusativeName,
		reservation.VkID)

	forward, err := vk.ForwardParam(
		reservation.VkID,
		[]int{reservation.Introduction})
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
		validate.ConfirmReservation,
		nil,
	)

	return "details", []*form.Field{field}
}
