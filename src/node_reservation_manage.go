package main

import (
	"ask-bot/src/ask"
	"ask-bot/src/vk"
	"errors"
	"fmt"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type ReservationManageNode struct {
	details *ask.ReservationDetail
}

func (node *ReservationManageNode) ID() string {
	return "reservation_cancel"
}

func (node *ReservationManageNode) Entry(user *User, c *Controls) error {
	details, err := c.Ask.ReservationsDetailsByVkID(user.id)
	if err != nil {
		return err
	}

	if details == nil {
		err = errors.New("there is no reservations")
		return zaperr.Wrap(err, "",
			zap.Int("user", user.id))
	}

	node.details = details

	buttons := [][]vk.Button{
		{
			{
				Label:   fmt.Sprintf("Отменить бронь на %s", node.details.AccusativeName),
				Color:   vk.SecondaryColor,
				Command: "cancel",
			},
		},
		{
			{
				Label:   "Назад",
				Color:   vk.NegativeColor,
				Command: "back",
			},
		},
	}

	// TO-DO: under consideration while
	message := fmt.Sprintf("У вас есть бронь на %s до %s.\nСтатус: %s",
		node.details.AccusativeName,
		node.details.Deadline.Time,
		node.details.Status)

	_, err = c.Vk.SendMessage(
		user.id,
		message,
		vk.CreateKeyboard(node.ID(), buttons),
		nil)
	return err
}

func (node *ReservationManageNode) NewMessage(user *User, c *Controls, message *vk.Message) (StateNode, bool, error) {
	return nil, false, nil
}

func (node *ReservationManageNode) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (StateNode, bool, error) {
	switch payload.Command {
	case "cancel":
		message := fmt.Sprintf("Вы уверены, что хотите отменить бронь на %s?",
			node.details.AccusativeName)
		return NewConfirmationNode(message, nil), false, nil

	case "back":
		return nil, true, nil
	}

	return nil, false, nil
}

func (node *ReservationManageNode) Back(user *User, c *Controls, prev_state StateNode) (bool, error) {
	confirmation, ok := prev_state.(*ConfirmationNode)
	if !ok {
		return false, node.Entry(user, c)
	}

	if confirmation.Answer {
		err := c.Ask.DeleteReservation(node.details.Id)
		if err != nil {
			return false, err
		}

		_, err = c.Vk.SendMessage(user.id, "Ваша бронь была успешно отменена.", "", nil)
		return true, err
	}

	return false, node.Entry(user, c)
}
