package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type AdminReservationNode struct {
	reservation *Reservation
	paginator   *Paginator[Reservation]
}

func (node *AdminReservationNode) ID() string {
	return "admin_reservation"
}

func (node *AdminReservationNode) Entry(user *User, c *Controls) error {
	reservations, err := c.Ask.UnderConsiderationReservations()
	if err != nil {
		return err
	}

	to_label := func(r Reservation) string {
		return r.Role
	}

	to_value := func(r Reservation) string {
		return strconv.Itoa(r.Id)
	}

	node.paginator = NewPaginator[Reservation](reservations,
		"reservations",
		RowsCount,
		ColsCount,
		to_label,
		to_value)

	return c.Vk.ChangeKeyboard(user.id,
		CreateKeyboard(node, node.paginator.Buttons()))
}

func (node *AdminReservationNode) NewMessage(user *User, c *Controls, message *Message) (StateNode, bool, error) {
	return nil, false, nil
}

func (node *AdminReservationNode) KeyboardEvent(user *User, c *Controls, payload *CallbackPayload) (StateNode, bool, error) {
	switch payload.Command {
	case "reservations":
		reservation, err := node.paginator.Object(payload.Value)
		if err != nil {
			return nil, false, err
		}

		node.reservation = reservation
		text := fmt.Sprintf("Роль: %s\nСтраница: https://vk.com/id%d\n",
			reservation.Role,
			reservation.VkID)

		forward, err := ForwardParam(
			reservation.VkID,
			[]int{reservation.Info})
		if err != nil {
			return nil, false, err
		}

		request := &MessageParams{
			Text: text,
			Params: api.Params{
				"forward": forward,
			},
		}

		field := NewConfirmReservation(request)
		form, err := NewForm(field)
		if err != nil {
			return nil, false, err
		}

		return &FormNode{
			Form: form,
		}, false, nil

	case "paginator":
		back := node.paginator.Control(payload.Value)

		if back {
			return nil, true, nil
		}

		return nil, false, c.Vk.ChangeKeyboard(user.id,
			CreateKeyboard(node, node.paginator.Buttons()))
	}
	return nil, false, nil
}

func (node *AdminReservationNode) Back(user *User, c *Controls, prev_state StateNode) (bool, error) {
	form, ok := prev_state.(*FormNode)
	if !ok {
		return false, nil
	}

	if !form.FilledOut {
		return false, node.Entry(user, c)
	}

	value, err := form.Form.Value(0)
	if err != nil {
		err := errors.New("form is not fullfilled")
		return false, zaperr.Wrap(err, "",
			zap.Any("form", form.Form))
	}

	confirm, err := ConvertValue[bool](value)
	if err != nil {
		return false, err
	}

	var message string
	if confirm {
		deadline, err := c.Ask.ConfirmReservation(node.reservation.Id)
		if err != nil {
			return false, err
		}

		message = fmt.Sprintf("Ваша бронь на %s успешно подтверждена! Вам нужно отрисовать приветствие до %s.",
			node.reservation.Role,
			deadline)
	} else {
		err := c.Ask.DeleteReservation(node.reservation.Id)
		if err != nil {
			return false, err
		}

		message = fmt.Sprintf("Ваша бронь на %s, к сожалению, отклонена. Попробуйте еще раз позже!",
			node.reservation.Role)
	}

	// notify user
	notification := &MessageParams{
		Id:   node.reservation.VkID,
		Text: message,
	}
	c.Notify <- notification

	return false, node.Entry(user, c)
}
