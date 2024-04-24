package states

import (
	"ask-bot/src/form"
	"ask-bot/src/paginator"
	"ask-bot/src/vk"
	"errors"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type Init struct {
	Silent bool

	paginator *paginator.Paginator[form.Option]
}

func (state *Init) ID() string {
	return "init"
}

func (state *Init) options(user *User, c *Controls) ([]form.Option, error) {
	options := []form.Option{}

	reservation, err := c.Ask.ReservationByVkID(user.Id)
	if err != nil {
		return nil, err
	}

	// same  id because i want to mask the difference between them
	if reservation == nil {
		options = append(options, form.Option{
			ID:    "reservation",
			Label: "Бронь",
			Value: &ReservationNew{},
		})
	} else {
		options = append(options, form.Option{
			ID:    "reservation",
			Label: "Бронь",
			Value: &ReservationManage{},
		})
	}

	options = append(options,
		form.Option{
			ID:    (&Points{}).ID(),
			Label: "Баллы",
			Value: &Points{},
		},
		form.Option{
			ID:    (&FAQ{}).ID(),
			Label: "FAQ",
			Value: &FAQ{},
		})

	is_admin, err := c.Ask.IsAdmin(user.Id)
	if err != nil {
		return nil, err
	}
	if is_admin {
		options = append(options, form.Option{
			ID:    (&Admin{}).ID(),
			Label: "Админ",
			Color: vk.PrimaryColor,
			Value: &Admin{},
		})
	}

	return options, nil
}

func (state *Init) updatePaginator(user *User, c *Controls) error {
	options, err := state.options(user, c)
	if err != nil {
		return err
	}

	if state.paginator == nil {
		config := &paginator.Config[form.Option]{
			Command:     "options",
			WithoutBack: true,

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

func (state *Init) Entry(user *User, c *Controls) error {
	err := state.updatePaginator(user, c)
	if err != nil {
		return err
	}

	if state.Silent {
		return c.Vk.ChangeKeyboard(user.Id,
			vk.CreateKeyboard(state.ID(), state.paginator.Buttons()))
	}

	_, err = c.Vk.SendMessage(user.Id,
		"Здравствуйте!",
		vk.CreateKeyboard(state.ID(), state.paginator.Buttons()),
		nil)
	return err
}

func (state *Init) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	return nil, state.Entry(user, c)
}

func (state *Init) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
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
		state.paginator.Control(payload.Value)
	}

	return nil, nil
}

func (state *Init) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	state.Silent = true

	return nil, state.Entry(user, c)
}
