package main

import (
	"ask-bot/src/ask"
	"ask-bot/src/dict"
	"ask-bot/src/form"
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
	return "reservation_manage"
}

func (node *ReservationManageNode) buttons() [][]vk.Button {
	actions := []vk.Button{}

	if node.details.Status == ask.ReservationStatuses.InProgress {
		actions = append(actions, vk.Button{
			Label:   "Прислать приветствие",
			Color:   vk.PrimaryColor,
			Command: "greeting",
		})
	}

	if node.details.Status != ask.ReservationStatuses.Poll {
		actions = append(actions, vk.Button{
			Label:   "Отменить бронь",
			Color:   vk.SecondaryColor,
			Command: "cancel",
		})
	}

	return [][]vk.Button{
		actions,
		{
			{
				Label:   "Назад",
				Color:   vk.NegativeColor,
				Command: "back",
			},
		},
	}
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

	var message string

	switch node.details.Status {
	case ask.ReservationStatuses.UnderConsideration:
		message = fmt.Sprintf("У вас есть бронь на %s на рассмотрении. Когда ее рассмотрят, вам придет сообщение.",
			node.details.AccusativeName)

	case ask.ReservationStatuses.InProgress:
		message = fmt.Sprintf("У вас есть бронь на %s до %s.",
			node.details.AccusativeName,
			node.details.Deadline.Time)

	case ask.ReservationStatuses.Done:
		message = fmt.Sprintf("Мы получили ваше приветствие на %s! Скоро будет создан опрос.",
			node.details.AccusativeName)

	case ask.ReservationStatuses.Poll:
		message = fmt.Sprintf("Ваше приветствие на %s уже на голосовании!",
			node.details.AccusativeName)
	}

	_, err = c.Vk.SendMessage(
		user.id,
		message,
		vk.CreateKeyboard(node.ID(), node.buttons()),
		nil)
	return err
}

func (node *ReservationManageNode) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	return nil, nil
}

func (node *ReservationManageNode) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "greeting":
		request := &vk.MessageParams{
			Text:   "Пришлите свое приветствие.",
			Params: nil,
		}

		field := form.NewField(
			"greeting",
			request,
			nil,
			ExtractAttachments,
			GreetingValidate,
			nil,
		)

		return NewActionNext(NewFormNode("greeting", nil, field)), nil

	case "cancel":
		message := &vk.MessageParams{
			Text: fmt.Sprintf("Вы уверены, что хотите отменить бронь на %s?",
				node.details.AccusativeName),
		}
		return NewActionNext(NewConfirmationNode("cancel", message)), nil

	case "back":
		return NewActionExit(nil), nil
	}

	return nil, nil
}

func (node *ReservationManageNode) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	if info == nil {
		return nil, node.Entry(user, c)
	}

	switch info.Payload {
	case "greeting":
		if node.details == nil {
			return nil, errors.New("no details in node")
		}

		greeting, err := dict.ExtractValue[string](info.Values, "greeting")
		if err != nil {
			return nil, err
		}

		err = c.Ask.CompleteReservation(node.details.Id, greeting)
		if err != nil {
			return nil, err
		}

	case "cancel":
		if node.details == nil {
			return nil, errors.New("no details in node")
		}

		answer, err := dict.ExtractValue[bool](info.Values, "confirmation")
		if err != nil {
			return nil, err
		}

		if answer {
			err := c.Ask.DeleteReservation(node.details.Id)
			if err != nil {
				return nil, err
			}

			_, err = c.Vk.SendMessage(user.id, "Ваша бронь была успешно отменена.", "", nil)
			return NewActionExit(nil), err
		}
	}

	return nil, node.Entry(user, c)
}
