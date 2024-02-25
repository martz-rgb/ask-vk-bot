package main

import (
	"ask-bot/src/ask"
	"ask-bot/src/dict"
	"ask-bot/src/form"
	"ask-bot/src/paginator"
	"ask-bot/src/vk"
	"errors"
	"fmt"
	"strconv"

	"github.com/SevereCloud/vksdk/v2/api"
)

type AdminReservationNode struct {
	reservation *ask.ReservationDetail
	paginator   *paginator.Paginator[ask.ReservationDetail]
}

func (node *AdminReservationNode) ID() string {
	return "admin_reservation"
}

// To-DO: print all reservations
func (node *AdminReservationNode) Entry(user *User, c *Controls) error {
	reservations, err := c.Ask.UnderConsiderationReservationsDetails()
	if err != nil {
		return err
	}

	to_label := func(r ask.ReservationDetail) string {
		return r.Name
	}

	to_value := func(r ask.ReservationDetail) string {
		return strconv.Itoa(r.Id)
	}

	node.paginator = paginator.New[ask.ReservationDetail](reservations,
		"reservations",
		paginator.DeafultRows,
		paginator.DefaultCols,
		false,
		to_label,
		to_value)

	return c.Vk.ChangeKeyboard(user.id,
		vk.CreateKeyboard(node.ID(), node.paginator.Buttons()))
}

func (node *AdminReservationNode) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	return nil, nil
}

func (node *AdminReservationNode) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "reservations":
		reservation, err := node.paginator.Object(payload.Value)
		if err != nil {
			return nil, err
		}

		node.reservation = reservation

		text := fmt.Sprintf("Роль: %s\nСтраница: https://vk.com/id%d\n",
			reservation.AccusativeName,
			reservation.VkID)

		forward, err := vk.ForwardParam(
			reservation.VkID,
			[]int{reservation.Info})
		if err != nil {
			return nil, err
		}

		field := form.NewField(
			"action",
			&vk.MessageParams{
				Text: text,
				Params: api.Params{
					"forward": forward,
				},
			},
			ConfirmReservationOptions,
			nil,
			ConfirmReservationValidate,
			nil,
		)
		return NewActionNext(NewFormNode("action", field)), nil

	case "paginator":
		back := node.paginator.Control(payload.Value)

		if back {
			return NewActionExit(&ExitInfo{}), nil
		}

		return nil, c.Vk.ChangeKeyboard(user.id,
			vk.CreateKeyboard(node.ID(), node.paginator.Buttons()))
	}
	return nil, nil
}

func (node *AdminReservationNode) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	if info == nil {
		return nil, node.Entry(user, c)
	}

	switch info.Payload {
	case "action":
		if node.reservation == nil {
			return nil, errors.New("no reservation in node")
		}
		values, err := dict.ExtractValue[dict.Dictionary](info.Values, "form")
		if err != nil {
			return nil, err
		}
		if values == nil {
			return nil, node.Entry(user, c)
		}

		action, err := dict.ExtractValue[bool](values, "action")
		if err != nil {
			return nil, err
		}

		var message, notification_message string
		if action == true {
			deadline, err := c.Ask.ConfirmReservation(node.reservation.Id)
			if err != nil {
				return nil, err
			}

			message = fmt.Sprintf("Бронь на %s была успешно подтверждена.",
				node.reservation.AccusativeName)
			notification_message = fmt.Sprintf("Ваша бронь на %s успешно подтверждена! Вам нужно отрисовать приветствие до %s.",
				node.reservation.AccusativeName,
				deadline)
		} else {
			err := c.Ask.DeleteReservation(node.reservation.Id)
			if err != nil {
				return nil, err
			}

			message = fmt.Sprintf("Бронь на %s была успешно удалена.",
				node.reservation.AccusativeName)
			notification_message = fmt.Sprintf("Ваша бронь на %s, к сожалению, отклонена. Попробуйте еще раз позже!",
				node.reservation.AccusativeName)
		}

		// notify user
		notification := &vk.MessageParams{
			Id:   node.reservation.VkID,
			Text: notification_message,
		}
		c.Notify <- notification

		_, err = c.Vk.SendMessage(user.id, message, "", nil)
		if err != nil {
			return nil, err
		}
	}

	return nil, node.Entry(user, c)
}
