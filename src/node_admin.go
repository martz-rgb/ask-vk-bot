package main

import (
	"ask-bot/src/form"
	"ask-bot/src/paginator"
	"ask-bot/src/vk"
	"errors"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type AdminNode struct {
	paginator *paginator.Paginator[form.Option]
}

func (node *AdminNode) ID() string {
	return "admin"
}

func (node *AdminNode) options() []form.Option {
	return []form.Option{
		{
			ID:    (&AdminReservationNode{}).ID(),
			Label: "Брони",
			Value: &AdminReservationNode{},
		},
		{
			ID:    (&RolesNode{}).ID(),
			Label: "Список ролей",
			Value: &RolesNode{},
		},
	}
}

func (node *AdminNode) updatePaginator() error {
	options := node.options()

	if node.paginator == nil {
		node.paginator = paginator.New[form.Option](
			options,
			"options",
			paginator.DeafultRows,
			paginator.DefaultCols,
			false,
			form.OptionToLabel,
			form.OptionToValue)
		return nil
	}
	node.paginator.ChangeObjects(options)

	return nil
}

func (node *AdminNode) Entry(user *User, c *Controls) error {
	ok, err := c.Ask.IsAdmin(user.id)
	if err != nil {
		return err
	}

	if !ok {
		err := errors.New("unathorized access")
		return zaperr.Wrap(err, "",
			zap.Any("user", user))
	}

	node.updatePaginator()

	return c.Vk.ChangeKeyboard(user.id, vk.CreateKeyboard(node.ID(), node.paginator.Buttons()))
}

func (node *AdminNode) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	return nil, nil
}
func (node *AdminNode) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "options":
		option, err := node.paginator.Object(payload.Value)
		if err != nil {
			return nil, err
		}

		next, ok := option.Value.(StateNode)
		if !ok {
			err := errors.New("failed to convert to StateNode")
			return nil, zaperr.Wrap(err, "",
				zap.Any("value", option.Value))
		}
		return NewActionNext(next), nil
	case "paginator":
		back := node.paginator.Control(payload.Value)

		if back {
			return NewActionExit(nil), nil
		}

		return nil, c.Vk.ChangeKeyboard(user.id,
			vk.CreateKeyboard(node.ID(), node.paginator.Buttons()))
	}

	return nil, nil
}

func (node *AdminNode) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	return nil, node.Entry(user, c)
}
