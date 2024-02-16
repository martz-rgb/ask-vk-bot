package main

type InitNode struct {
	Silent bool
}

func (node *InitNode) ID() string {
	return "init"
}

func (node *InitNode) buttons(is_admin bool) [][]Button {
	buttons := [][]Button{{
		{
			Label:   "Список ролей",
			Color:   "secondary",
			Command: (&RolesNode{}).ID(),
		},
		{
			Label:   "Бронь",
			Color:   "secondary",
			Command: (&ReservationNode{}).ID(),
		},
		{
			Label:   "Баллы",
			Color:   "secondary",
			Command: (&PointsNode{}).ID(),
		}},
		{{
			Label:   "Дедлайн",
			Color:   "secondary",
			Command: (&DeadlineNode{}).ID(),
		},
			{
				Label:   "FAQ",
				Color:   "secondary",
				Command: (&FAQNode{}).ID(),
			},
		}}

	if is_admin {
		buttons[len(buttons)-1] = append(buttons[len(buttons)-1],
			Button{
				Label:   "Админ",
				Color:   PrimaryColor,
				Command: (&AdminNode{}).ID(),
			},
		)
	}

	return buttons
}

func (node *InitNode) Entry(user *User, c *Controls) error {
	is_admin, err := c.Ask.IsAdmin(user.id)
	if err != nil {
		return err
	}

	if node.Silent {
		return c.Vk.ChangeKeyboard(user.id,
			CreateKeyboard(node, node.buttons(is_admin)))
	}

	_, err = c.Vk.SendMessage(user.id,
		"Здравствуйте!",
		CreateKeyboard(node, node.buttons(is_admin)),
		nil)
	return err
}

func (node *InitNode) NewMessage(user *User, c *Controls, message *Message) (StateNode, bool, error) {
	return nil, false, nil
}

func (node *InitNode) KeyboardEvent(user *User, c *Controls, payload *CallbackPayload) (StateNode, bool, error) {
	switch payload.Command {
	case (&FAQNode{}).ID():
		return &FAQNode{}, false, nil
	case (&RolesNode{}).ID():
		return &RolesNode{}, false, nil
	case (&ReservationNode{}).ID():
		return &ReservationNode{}, false, nil
	case (&PointsNode{}).ID():
		return &PointsNode{}, false, nil
	case (&DeadlineNode{}).ID():
		return &DeadlineNode{}, false, nil
	case (&AdminNode{}).ID():
		return &AdminNode{}, false, nil
	}

	return nil, false, nil
}

func (node *InitNode) Back(user *User, c *Controls, prev StateNode) (bool, error) {
	is_admin, err := c.Ask.IsAdmin(user.id)
	if err != nil {
		return false, err
	}

	return false, c.Vk.ChangeKeyboard(user.id, CreateKeyboard(node, node.buttons(is_admin)))
}
