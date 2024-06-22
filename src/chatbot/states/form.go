package states

import (
	"ask-bot/src/datatypes/form"
	"ask-bot/src/datatypes/form/check"
	"ask-bot/src/vk"
	"errors"

	"github.com/hori-ryota/zaperr"
)

type Form struct {
	f       *form.Form
	payload string
}

func NewForm(payload string,
	fields ...form.Field) (*Form, error) {
	f, err := form.NewForm(fields...)
	if err != nil {
		return nil, err
	}

	return &Form{
		f:       f,
		payload: payload,
	}, nil
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
		return NewActionExit(state.exitInfo()), nil
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
			return NewActionExit(state.exitInfo()), nil
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
	return nil, state.Entry(user, c)
}

func (state *Form) sendRequest(user *User, c *Controls) error {
	request := state.f.Request()

	_, err := c.Vk.SendMessageParams(
		user.Id,
		request.Message,
		vk.CreateKeyboard(state.ID(), state.f.Buttons()))
	return err
}

func (state *Form) set(user *User, c *Controls, input interface{}) (end bool, err error) {
	var info *check.Result

	switch value := input.(type) {
	case *vk.Message:
		info, err = state.f.SetFromMessage(value)
	case string:
		info, err = state.f.SetFromOption(value)
	}

	if err != nil {
		return false, err
	}

	if !info.Ok() {
		_, err = c.Vk.SendMessageParams(user.Id, info.ErrorToMessageParams(), "")
		return false, err
	}

	end, err = state.f.Next()
	if err != nil {
		return false, err
	}

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
