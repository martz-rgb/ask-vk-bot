package states

import (
	"ask-bot/src/ask"
	"ask-bot/src/chatbot/states/extract"
	"ask-bot/src/chatbot/states/validate"
	"ask-bot/src/dict"
	"ask-bot/src/form"
	"ask-bot/src/paginator"
	ts "ask-bot/src/templates"
	"ask-bot/src/vk"
	"errors"

	"github.com/SevereCloud/vksdk/v2/api"
)

type ReservationNew struct {
	paginator *paginator.Paginator[ask.Role]

	role *ask.Role
}

func (state *ReservationNew) ID() string {
	return "reservation"
}

func (state *ReservationNew) Entry(user *User, c *Controls) error {
	roles, err := c.Ask.AvailableRoles()
	if err != nil {
		return err
	}

	config := &paginator.Config[ask.Role]{
		Command: "roles",

		ToLabel: func(role ask.Role) string {
			return role.ShownName
		},
		ToValue: func(role ask.Role) string {
			return role.Name
		},
	}

	state.paginator = paginator.New(
		roles,
		config.MustBuild())

	message, err := ts.ParseTemplate(
		ts.MsgReservationNew,
		ts.MsgReservationNewData{},
	)
	if err != nil {
		return err
	}

	_, err = c.Vk.SendMessage(user.Id,
		message,
		vk.CreateKeyboard(state.ID(), state.paginator.Buttons()),
		nil)
	return err
}

func (state *ReservationNew) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	roles, err := c.Ask.AvailableRolesStartWith(message.Text)
	if err != nil {
		return nil, err
	}

	state.paginator.ChangeObjects(roles)

	return nil, c.Vk.ChangeKeyboard(user.Id, vk.CreateKeyboard(state.ID(), state.paginator.Buttons()))
}

func (state *ReservationNew) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "roles":
		role, err := state.paginator.Object(payload.Value)
		if err != nil {
			return nil, err
		}
		state.role = role

		message, err := ts.ParseTemplate(
			ts.MsgReservationNewConfirmation,
			ts.MsgReservationNewConfirmationData{
				Role: *role,
			})
		if err != nil {
			return nil, err
		}

		return NewActionNext(NewConfirmation(
			"confirmation",
			&vk.MessageParams{
				Text: message,
			})), nil

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

func (state *ReservationNew) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	if info == nil {
		return nil, state.Entry(user, c)
	}

	switch info.Payload {
	case "confirmation":
		answer, err := dict.ExtractValue[bool](info.Values, "confirmation")
		if err != nil {
			return nil, err
		}

		if answer {
			request, err := ts.ParseTemplate(
				ts.MsgReservationNewIntro,
				ts.MsgReservationNewConfirmationData{},
			)
			if err != nil {
				return nil, err
			}

			field := form.NewField(
				"about",
				&vk.MessageParams{
					Text: request,
				},
				nil,
				extract.ID,
				validate.InfoAbout,
				nil)

			return NewActionNext(NewForm("about", nil, field)), nil
		}

	case "about":
		if state.role == nil {
			return nil, errors.New("no role in state")
		}

		id, err := dict.ExtractValue[int](info.Values, "about")
		if err != nil {
			return nil, err
		}

		err = c.Ask.AddReservation(user.Id, state.role.Name, id)
		if err != nil {
			return nil, err
		}

		message, err := ts.ParseTemplate(
			ts.MsgReservationNewSuccess,
			ts.MsgReservationNewSuccessData{
				Role: *state.role,
			})
		if err != nil {
			return nil, err
		}
		forward, err := vk.ForwardParam(user.Id, []int{id})
		if err != nil {
			return nil, err
		}

		_, err = c.Vk.SendMessage(user.Id, message, "", api.Params{
			"forward": forward,
		})
		return NewActionExit(nil), err
	}

	return nil, state.Entry(user, c)
}
