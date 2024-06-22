package states

import (
	"ask-bot/src/datatypes/form"
	"ask-bot/src/datatypes/paginator"
	"ask-bot/src/vk"
	"errors"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type Admin struct {
	paginator *paginator.Paginator[form.Option]
}

func (state *Admin) ID() string {
	return "admin"
}

func (state *Admin) options() []form.Option {
	return []form.Option{
		{
			ID:    (&AdminReservation{}).ID(),
			Label: "Брони",
			Value: &AdminReservation{},
		},
		{
			ID:    (&RolesList{}).ID(),
			Label: "Список ролей",
			Value: &RolesList{},
		},
	}
}

func (state *Admin) updatePaginator() error {
	options := state.options()

	if state.paginator == nil {
		config := &paginator.Config[form.Option]{
			Command: "options",
			ToLabel: form.OptionToLabel,
			ToColor: form.OptionToColor,
			ToValue: form.OptionToValue,
		}

		state.paginator = paginator.New[form.Option](
			options,
			config.MustBuild())
		return nil
	}
	state.paginator.ChangeObjects(options)

	return nil
}

func (state *Admin) Entry(user *User, c *Controls) error {
	ok, err := c.Ask.IsAdmin(user.Id)
	if err != nil {
		return err
	}

	if !ok {
		err := errors.New("unathorized access")
		return zaperr.Wrap(err, "",
			zap.Any("user", user))
	}

	state.updatePaginator()

	return c.Vk.ChangeKeyboard(user.Id, vk.CreateKeyboard(state.ID(), state.paginator.Buttons()))
}

func (state *Admin) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	return nil, nil
}
func (state *Admin) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "options":
		option, err := state.paginator.Object(payload.Value)
		if err != nil {
			return nil, err
		}

		next, ok := option.Value.(State)
		if !ok {
			err := errors.New("failed to convert to StateNode")
			return nil, zaperr.Wrap(err, "",
				zap.Any("value", option.Value))
		}
		return NewActionNext(next), nil
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

func (state *Admin) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	return nil, state.Entry(user, c)
}
