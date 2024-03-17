package states

import (
	"ask-bot/src/dict"
	"ask-bot/src/form"
	"ask-bot/src/vk"
	"errors"

	"github.com/hori-ryota/zaperr"
)

type Form struct {
	f *form.Form

	payload      string
	confirmation func(values dict.Dictionary) (*vk.MessageParams, error)
}

func NewForm(payload string,
	confirmation func(values dict.Dictionary) (*vk.MessageParams, error),
	fields ...*form.Field) *Form {
	return &Form{
		f:            form.NewForm(fields...),
		confirmation: confirmation,

		payload: payload,
	}
}

func (state *Form) ID() string {
	return "form"
}

func (state *Form) Entry(user *User, c *Controls) error {
	if state.f == nil {
		err := errors.New("no form is provided")
		return zaperr.Wrap(err, "")
	}

	return state.sendRequest(user, c)
}

func (state *Form) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	end, err := state.set(user, c, message)
	if err != nil {
		return nil, err
	}

	if end {
		return state.endAction()
	}

	return nil, nil
}

func (state *Form) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "form":
		end, err := state.set(user, c, payload.Value)
		if err != nil {
			return nil, err
		}

		if end {
			return state.endAction()
		}

	case "paginator":
		back := state.f.Control(payload.Value)

		if back {
			return NewActionExit(nil), nil
		}

		return nil, c.Vk.ChangeKeyboard(user.Id,
			vk.CreateKeyboard(state.ID(), state.f.Buttons()))
	}

	return nil, nil
}

func (state *Form) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	if info == nil {
		return nil, state.Entry(user, c)
	}

	switch info.Payload {
	case "confirm":
		answer, err := dict.ExtractValue[bool](info.Values, "confirmation")
		if err != nil {
			return nil, err
		}

		if answer {
			return NewActionExit(state.exitInfo()), nil
		} else {
			return NewActionExit(nil), nil
		}
	}

	return nil, state.Entry(user, c)
}

func (state *Form) sendRequest(user *User, c *Controls) error {
	request := state.f.Request()

	_, err := c.Vk.SendMessageParams(
		user.Id,
		request,
		vk.CreateKeyboard(state.ID(), state.f.Buttons()))
	return err
}

func (state *Form) set(user *User, c *Controls, input interface{}) (end bool, err error) {
	var info *vk.MessageParams

	switch value := input.(type) {
	case *vk.Message:
		info, err = state.f.SetFromMessage(value)
	case string:
		info, err = state.f.SetFromOption(value)
	}

	if err != nil {
		return false, err
	}

	if info != nil {
		_, err = c.Vk.SendMessage(user.Id, info.Text, "", info.Params)
		return false, err
	}

	end = state.f.Next()
	if !end {
		return false, state.sendRequest(user, c)
	}
	return true, nil
}

func (state *Form) exitInfo() *ExitInfo {
	return &ExitInfo{
		Values:  state.f.Values(),
		Payload: state.payload,
	}
}

func (state *Form) endAction() (*Action, error) {
	if state.confirmation == nil {
		return NewActionExit(state.exitInfo()), nil
	}

	message, err := state.confirmation(state.f.Values())
	if err != nil {
		return nil, err
	}

	return NewActionNext(
		NewConfirmation("confirm", message),
	), nil
}
