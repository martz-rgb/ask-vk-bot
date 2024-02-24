package main

import (
	"ask-bot/src/form"
	"ask-bot/src/paginator"
	"ask-bot/src/vk"
	"errors"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type InitNode struct {
	Silent bool

	paginator *paginator.Paginator[form.Option]
}

func (node *InitNode) ID() string {
	return "init"
}

func (node *InitNode) options(user *User, c *Controls) ([]form.Option, error) {
	options := []form.Option{}

	reservations, err := c.Ask.ReservationsByVkID(user.id)
	if err != nil {
		return nil, err
	}
	if len(reservations) == 0 {
		options = append(options, form.Option{
			ID:    (&ReservationNode{}).ID(),
			Label: "Бронь",
			Value: &ReservationNode{},
		})
	}

	options = append(options,
		form.Option{
			ID:    (&PointsNode{}).ID(),
			Label: "Баллы",
			Value: &PointsNode{},
		},
		form.Option{
			ID:    (&FAQNode{}).ID(),
			Label: "FAQ",
			Value: &FAQNode{},
		})

	is_admin, err := c.Ask.IsAdmin(user.id)
	if err != nil {
		return nil, err
	}
	if is_admin {
		options = append(options, form.Option{
			ID:    (&AdminNode{}).ID(),
			Label: "Админ",
			Value: &AdminNode{},
		})
	}

	return options, nil
}

func (node *InitNode) updatePaginator(user *User, c *Controls) error {
	options, err := node.options(user, c)
	if err != nil {
		return err
	}

	if node.paginator == nil {
		node.paginator = paginator.New[form.Option](
			options,
			"options",
			paginator.DeafultRows,
			paginator.DefaultCols,
			true,
			form.OptionToLabel,
			form.OptionToValue)
		return nil
	}
	node.paginator.ChangeObjects(options)

	return nil

}

func (node *InitNode) Entry(user *User, c *Controls) error {
	err := node.updatePaginator(user, c)
	if err != nil {
		return err
	}

	if node.Silent {
		return c.Vk.ChangeKeyboard(user.id,
			vk.CreateKeyboard(node.ID(), node.paginator.Buttons()))
	}

	_, err = c.Vk.SendMessage(user.id,
		"Здравствуйте!",
		vk.CreateKeyboard(node.ID(), node.paginator.Buttons()),
		nil)
	return err
}

func (node *InitNode) NewMessage(user *User, c *Controls, message *vk.Message) (StateNode, bool, error) {
	return nil, false, nil
}

func (node *InitNode) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (StateNode, bool, error) {
	switch payload.Command {
	case "options":
		option, err := node.paginator.Object(payload.Value)
		if err != nil {
			return nil, false, err
		}

		next, ok := option.Value.(StateNode)
		if !ok {
			err := errors.New("failed to convert to StateNode")
			return nil, false, zaperr.Wrap(err, "",
				zap.Any("value", option.Value))
		}
		return next, false, nil
	case "paginator":
		node.paginator.Control(payload.Value)
	}

	return nil, false, nil
}

func (node *InitNode) Back(user *User, c *Controls, prev StateNode) (bool, error) {
	err := node.updatePaginator(user, c)
	if err != nil {
		return false, err
	}

	return false, c.Vk.ChangeKeyboard(user.id, vk.CreateKeyboard(node.ID(), node.paginator.Buttons()))
}
