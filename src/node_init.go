package main

type InitNode struct{}

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

func (node *InitNode) Entry(user *User, ask *Ask, vk *VK) error {
	is_admin, err := ask.IsAdmin(user.id)

	_, err = vk.SendMessage(user.id, "Здравствуйте!", CreateKeyboard(node, node.buttons(is_admin)), nil)
	return err
}

func (node *InitNode) NewMessage(user *User, ask *Ask, vk *VK, message *Message) (StateNode, bool, error) {
	return nil, false, nil
}

func (node *InitNode) KeyboardEvent(user *User, ask *Ask, vk *VK, payload *CallbackPayload) (StateNode, bool, error) {
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

func (node *InitNode) Back(user *User, ask *Ask, vk *VK, prev StateNode) (bool, error) {
	is_admin, err := ask.IsAdmin(user.id)
	if err != nil {
		return false, err
	}

	return false, vk.ChangeKeyboard(user.id, CreateKeyboard(node, node.buttons(is_admin)))
}
