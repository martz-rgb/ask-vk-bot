package states

import (
	"ask-bot/src/ask"
	"ask-bot/src/datatypes/dict"
	"ask-bot/src/datatypes/form"
	"ask-bot/src/datatypes/form/check"
	"ask-bot/src/datatypes/form/extrude"
	"ask-bot/src/datatypes/paginator"
	ts "ask-bot/src/templates"
	"ask-bot/src/vk"
	"errors"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type ReservationManage struct {
	paginator   *paginator.Paginator[form.Option]
	reservation ask.Reservation
}

func (state *ReservationManage) ID() string {
	return "reservation_manage"
}

func (state *ReservationManage) options() (options []form.Option) {
	if state.reservation.Status == ask.ReservationStatuses.InProgress {
		options = append(options, form.Option{
			ID:    "greeting",
			Label: "Приветствие",
			Color: vk.PrimaryColor,
		})
	}

	if state.reservation.Status != ask.ReservationStatuses.Poll {
		options = append(options, form.Option{
			ID:    "cancel",
			Label: "Отменить",
			Color: vk.SecondaryColor,
		})
	}

	return
}

func (state *ReservationManage) Entry(user *User, c *Controls) error {
	reservation, err := c.Ask.ReservationByVkID(user.Id)
	if err != nil {
		return err
	}

	if reservation == nil {
		err = errors.New("there is no reservations")
		return zaperr.Wrap(err, "",
			zap.Int("user", user.Id))
	}

	state.reservation = *reservation

	var message string

	switch state.reservation.Status {
	case ask.ReservationStatuses.UnderConsideration:
		message, err = ts.ParseTemplate(
			ts.MsgReservationUnderConsideration,
			ts.MsgReservationUnderConsiderationData{
				Reservation: state.reservation,
			},
		)

	case ask.ReservationStatuses.InProgress:
		message, err = ts.ParseTemplate(
			ts.MsgReservationInProgress,
			ts.MsgReservationInProgressData{
				Reservation: state.reservation,
			},
		)

	case ask.ReservationStatuses.Done:
		// TO-DO: try to get info about postponed poll from postponed
		message, err = ts.ParseTemplate(
			ts.MsgReservationDone,
			ts.MsgReservationDoneData{
				Reservation: state.reservation,
			},
		)

	case ask.ReservationStatuses.Poll:
		message, err = ts.ParseTemplate(
			ts.MsgReservationPoll,
			ts.MsgReservationPollData{
				Reservation: state.reservation,
				Link:        c.Vk.PostLink(int(reservation.Poll.Int32)),
			},
		)
	}

	if err != nil {
		return err
	}

	config := &paginator.Config[form.Option]{
		Command: "options",

		ToLabel: form.OptionToLabel,
		ToColor: form.OptionToColor,
		ToValue: form.OptionToValue,
	}

	state.paginator = paginator.New(
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
			message, err := ts.ParseTemplate(
				ts.MsgReservationGreetingRequest,
				ts.MsgReservationGreetingRequestData{})
			if err != nil {
				return nil, err
			}

			image := form.Field{
				Name:           "image",
				BuildRequest:   form.AlwaysRequest(&vk.MessageParams{Text: message}, nil),
				ExtrudeMessage: extrude.Images,
				Check:          check.NotEmpty,
			}

			form, err := NewForm("greeting", image)
			return NewActionNext(form), err

		case "cancel":
			message, err := ts.ParseTemplate(
				ts.MsgReservationCancel,
				ts.MsgReservationCancelData{
					Reservation: state.reservation,
				})
			if err != nil {
				return nil, err
			}

			confirmation := form.Field{
				Name:           "confirmation",
				BuildRequest:   form.AlwaysConfirm(&vk.MessageParams{Text: message}),
				ExtrudeMessage: nil,
				Check:          check.NotEmptyBool,
			}

			form, err := NewForm("cancel", confirmation)
			return NewActionNext(form), err
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
		greeting, err := dict.ExtractStruct[struct {
			Image string
		}](info.Values)
		if err != nil {
			return nil, err
		}

		err = c.Ask.CompleteReservation(state.reservation.VkID, greeting.Image)
		if err != nil {
			return nil, err
		}

	case "cancel":
		data, err := dict.ExtractStruct[struct {
			Confirmation bool
		}](info.Values)
		if err != nil {
			return nil, err
		}

		if !data.Confirmation {
			return nil, state.Entry(user, c)
		}

		err = c.Ask.DeleteReservation(state.reservation.VkID)
		if err != nil {
			return nil, err
		}

		message, err := ts.ParseTemplate(
			ts.MsgReservationCancelSuccess,
			ts.MsgReservationCancelSuccessData{},
		)
		if err != nil {
			return nil, err
		}

		_, err = c.Vk.SendMessage(user.Id, message, "", nil)
		return NewActionExit(nil), err
	}

	return nil, state.Entry(user, c)
}
