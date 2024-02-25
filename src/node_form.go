package main

import (
	"ask-bot/src/form"
	"ask-bot/src/vk"
	"errors"

	"github.com/hori-ryota/zaperr"
)

type FormNode struct {
	f *form.Form

	payload string
}

func NewFormNode(payload string, fields ...*form.Field) *FormNode {
	return &FormNode{
		f: form.NewForm(fields...),

		payload: payload,
	}
}

func (node *FormNode) ID() string {
	return "form"
}

func (node *FormNode) Entry(user *User, c *Controls) error {
	if node.f == nil {
		err := errors.New("no form is provided")
		return zaperr.Wrap(err, "")
	}

	return node.sendRequest(user, c)
}

func (node *FormNode) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	end, err := node.set(user, c, message)
	if err != nil {
		return nil, err
	}

	if end {
		return node.exit_action(), nil
	}

	return nil, nil
}

func (node *FormNode) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "form":
		end, err := node.set(user, c, payload.Value)
		if err != nil {
			return nil, err
		}

		if end {
			return node.exit_action(), nil
		}

	case "paginator":
		back := node.f.Control(payload.Value)

		if back {
			return node.exit_action(), nil
		}

		return nil, c.Vk.ChangeKeyboard(user.id,
			vk.CreateKeyboard(node.ID(), node.f.Buttons()))
	}

	return nil, nil
}

func (node *FormNode) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	return nil, node.Entry(user, c)
}

func (node *FormNode) sendRequest(user *User, c *Controls) error {
	request := node.f.Request()

	_, err := c.Vk.SendMessageParams(
		user.id,
		request,
		vk.CreateKeyboard(node.ID(), node.f.Buttons()))
	return err
}

func (node *FormNode) set(user *User, c *Controls, input interface{}) (end bool, err error) {
	var info *vk.MessageParams

	switch value := input.(type) {
	case *vk.Message:
		info, err = node.f.SetFromMessage(value)
	case string:
		info, err = node.f.SetFromOption(value)
	}

	if err != nil {
		return false, err
	}

	if info != nil {
		_, err = c.Vk.SendMessage(user.id, info.Text, "", info.Params)
		return false, err
	}

	end = node.f.Next()
	if !end {
		return false, node.sendRequest(user, c)
	}
	return true, nil
}

func (node *FormNode) exit_action() *Action {
	return NewActionExit(&ExitInfo{
		Values: map[string]interface{}{
			"form": node.f.Values(),
		},
		Payload: node.payload,
	})
}
