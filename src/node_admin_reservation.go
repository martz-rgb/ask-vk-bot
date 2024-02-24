package main

import (
	"ask-bot/src/ask"
	"ask-bot/src/form"
	"ask-bot/src/paginator"
	"ask-bot/src/vk"
	"fmt"
	"strconv"

	"github.com/SevereCloud/vksdk/v2/api"
)

type AdminReservationNode struct {
	reservation *ask.Reservation
	paginator   *paginator.Paginator[ask.Reservation]
}

func (node *AdminReservationNode) ID() string {
	return "admin_reservation"
}

func (node *AdminReservationNode) Entry(user *User, c *Controls) error {
	reservations, err := c.Ask.UnderConsiderationReservations()
	if err != nil {
		return err
	}

	to_label := func(r ask.Reservation) string {
		return r.Role
	}

	to_value := func(r ask.Reservation) string {
		return strconv.Itoa(r.Id)
	}

	node.paginator = paginator.New[ask.Reservation](reservations,
		"reservations",
		paginator.DeafultRows,
		paginator.DefaultCols,
		false,
		to_label,
		to_value)

	return c.Vk.ChangeKeyboard(user.id,
		vk.CreateKeyboard(node.ID(), node.paginator.Buttons()))
}

func (node *AdminReservationNode) NewMessage(user *User, c *Controls, message *vk.Message) (StateNode, bool, error) {
	return nil, false, nil
}

func (node *AdminReservationNode) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (StateNode, bool, error) {
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

		forward, err := vk.ForwardParam(
			reservation.VkID,
			[]int{reservation.Info})
		if err != nil {
			return nil, false, err
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
		return NewFormNode(field), false, nil

	case "paginator":
		back := node.paginator.Control(payload.Value)

		if back {
			return nil, true, nil
		}

		return nil, false, c.Vk.ChangeKeyboard(user.id,
			vk.CreateKeyboard(node.ID(), node.paginator.Buttons()))
	}
	return nil, false, nil
}

func (node *AdminReservationNode) Back(user *User, c *Controls, prev_state StateNode) (bool, error) {
	form_node, ok := prev_state.(*FormNode)
	if !ok {
		return false, node.Entry(user, c)
	}

	if !form_node.IsFilled() {
		return false, node.Entry(user, c)
	}

	values := form_node.Values()
	action, err := form.ExtractValue[bool](values, "action")
	if err != nil {
		return false, err
	}

	var message string
	if action == true {
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
	notification := &vk.MessageParams{
		Id:   node.reservation.VkID,
		Text: message,
	}
	c.Notify <- notification

	return false, node.Entry(user, c)
}
