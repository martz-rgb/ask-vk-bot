package states

import (
	"ask-bot/src/ask"
	"ask-bot/src/datatypes/paginator"
	ts "ask-bot/src/templates"
	"ask-bot/src/vk"
)

type RolesList struct {
	paginator *paginator.Paginator[ask.Role]
}

func (state *RolesList) ID() string {
	return "roles"
}

func (state *RolesList) Entry(user *User, c *Controls) error {
	roles, err := c.Ask.Roles()
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
		ts.MsgAdminRoles,
		ts.MsgAdminRolesData{},
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

func (state *RolesList) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	roles, err := c.Ask.RolesStartWith(message.Text)
	if err != nil {
		return nil, err
	}

	state.paginator.ChangeObjects(roles)

	return nil, c.Vk.ChangeKeyboard(user.Id, vk.CreateKeyboard(state.ID(), state.paginator.Buttons()))
}

func (state *RolesList) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "roles":
		role, err := state.paginator.Object(payload.Value)
		if err != nil {
			return nil, err
		}

		message, err := ts.ParseTemplate(
			ts.MsgAdminRolesItem,
			ts.MsgAdminRolesItemData{
				Role: *role,
			})
		if err != nil {
			return nil, err
		}

		_, err = c.Vk.SendMessage(user.Id, message, "", nil)
		return nil, err
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

func (state *RolesList) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	return nil, state.Entry(user, c)
}
