package states

import (
	"ask-bot/src/ask"
	"ask-bot/src/chatbot/states/extract"
	"ask-bot/src/chatbot/states/validate"
	"ask-bot/src/dict"
	"ask-bot/src/form"
	"ask-bot/src/paginator"
	"ask-bot/src/vk"
	"errors"
	"fmt"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type ReservationManage struct {
	paginator *paginator.Paginator[form.Option]
	details   *ask.Reservation
}

func (state *ReservationManage) ID() string {
	return "reservation_manage"
}

func (state *ReservationManage) options() (options []form.Option) {
	if state.details.Status == ask.ReservationStatuses.InProgress {
		options = append(options, form.Option{
			ID:    "greeting",
			Label: "Приветствие",
			Color: vk.PrimaryColor,
		})
	}

	if state.details.Status != ask.ReservationStatuses.Poll {
		options = append(options, form.Option{
			ID:    "cancel",
			Label: "Отменить",
			Color: vk.SecondaryColor,
		})
	}

	return
}

func (state *ReservationManage) Entry(user *User, c *Controls) error {
	details, err := c.Ask.ReservationByVkID(user.Id)
	if err != nil {
		return err
	}

	if details == nil {
		err = errors.New("there is no reservations")
		return zaperr.Wrap(err, "",
			zap.Int("user", user.Id))
	}

	state.details = details

	var message string

	switch state.details.Status {
	case ask.ReservationStatuses.UnderConsideration:
		message = fmt.Sprintf("У вас есть бронь на %s на рассмотрении. Когда ее рассмотрят, вам придет сообщение.",
			state.details.AccusativeName)

	case ask.ReservationStatuses.InProgress:
		message = fmt.Sprintf("У вас есть бронь на %s до %s.",
			state.details.AccusativeName,
			state.details.Deadline.Time)

	case ask.ReservationStatuses.Done:
		// TO-DO: try to get info about postponed poll from postponed
		message = fmt.Sprintf("Мы получили ваше приветствие на %s! Скоро будет создан опрос.",
			state.details.AccusativeName)

	case ask.ReservationStatuses.Poll:
		message = fmt.Sprintf("Опрос начался! Посмотреть на него можно здесь: https://vk.com/wall%d_%d",
			c.Vk.ID(), details.Poll.Int32)
	}

	config := &paginator.Config[form.Option]{
		Command: "options",

		ToLabel: form.OptionToLabel,
		ToColor: form.OptionToColor,
		ToValue: form.OptionToValue,
	}

	state.paginator = paginator.New[form.Option](
		state.options(),
		config.MustBuild())

	_, err = c.Vk.SendMessage(
		user.Id,
		message,
		vk.CreateKeyboard(state.ID(), state.paginator.Buttons()),
		nil)
	return err
}

func (state *ReservationManage) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	return nil, nil
}

func (state *ReservationManage) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "options":
		option, err := state.paginator.Object(payload.Value)
		if err != nil {
			return nil, err
		}

		switch option.ID {
		case "greeting":
			request := &vk.MessageParams{
				Text:   "Пришлите свое приветствие.",
				Params: nil,
			}

			field := form.NewField(
				"greeting",
				request,
				nil,
				extract.Images,
				validate.NotEmpty,
				nil,
			)

			return NewActionNext(NewForm("greeting", nil, field)), nil

		case "cancel":
			message := &vk.MessageParams{
				Text: fmt.Sprintf("Вы уверены, что хотите отменить бронь на %s?",
					state.details.AccusativeName),
			}
			return NewActionNext(NewConfirmation("cancel", message)), nil
		}
	case "paginator":
		back := state.paginator.Control(payload.Value)

		if back {
			return NewActionExit(nil), nil
		}

		return nil, c.Vk.ChangeKeyboard(user.Id,
			vk.CreateKeyboard(state.ID(), state.paginator.Buttons()))
	}

	return nil, nil
}

// TO-DO resend greeting maybe?
func (state *ReservationManage) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	if info == nil {
		return nil, state.Entry(user, c)
	}

	switch info.Payload {
	case "greeting":
		if state.details == nil {
			return nil, errors.New("no details in state")
		}

		greeting, err := dict.ExtractValue[string](info.Values, "greeting")
		if err != nil {
			return nil, err
		}

		err = c.Ask.CompleteReservation(state.details.VkID, greeting)
		if err != nil {
			return nil, err
		}

	case "cancel":
		if state.details == nil {
			return nil, errors.New("no details in state")
		}

		answer, err := dict.ExtractValue[bool](info.Values, "confirmation")
		if err != nil {
			return nil, err
		}

		if answer {
			err := c.Ask.DeleteReservation(state.details.VkID)
			if err != nil {
				return nil, err
			}

			_, err = c.Vk.SendMessage(user.Id, "Ваша бронь была успешно отменена.", "", nil)
			return NewActionExit(nil), err
		}
	}

	return nil, state.Entry(user, c)
}
