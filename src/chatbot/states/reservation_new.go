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

		confirm_msg, err := ts.ParseTemplate(
			ts.MsgReservationNewConfirmation,
			ts.MsgReservationNewConfirmationData{
				Role: *role,
			})
		if err != nil {
			return nil, err
		}

		introduction_msg, err := ts.ParseTemplate(
			ts.MsgReservationNewIntro,
			ts.MsgReservationNewIntroData{},
		)
		if err != nil {
			return nil, err
		}

		confirm := form.Field{
			Name:           "confirmation",
			BuildRequest:   form.AlwaysConfirm(&vk.MessageParams{Text: confirm_msg}),
			ExtrudeMessage: nil,
			Check:          check.NotEmptyBool,
		}

		introduction := form.Field{
			Name: "introduction",
			BuildRequest: func(d dict.Dictionary) (*form.Request, bool, error) {
				data, err := dict.ExtractStruct[struct {
					Confirmation bool
				}](d)
				if err != nil {
					return nil, false, err
				}

				if !data.Confirmation {
					return nil, true, nil
				}

				return &form.Request{
					Message: &vk.MessageParams{
						Text: introduction_msg,
					},
				}, false, nil
			},
			ExtrudeMessage: extrude.ID,
			Check:          check.NotEmptyPositiveInt,
		}

		form, err := NewForm("introduction", confirm, introduction)
		return NewActionNext(form), err

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
	case "introduction":
		data, err := dict.ExtractStruct[struct {
			Confirmation bool
			Introduction int
		}](info.Values)
		if err != nil {
			return nil, err
		}

		if !data.Confirmation {
			return nil, state.Entry(user, c)
		}

		err = c.Ask.AddReservation(user.Id, state.role.Name, data.Introduction)
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

		forward, err := vk.ForwardParam(user.Id, []int{data.Introduction})
		if err != nil {
			return nil, err
		}

		_, err = c.Vk.SendMessage(user.Id, message, "", forward)
		return NewActionExit(nil), err
	}

	return nil, state.Entry(user, c)
}
