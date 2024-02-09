package main

type InitNode struct{}

func (node *InitNode) ID() string {
	return "init"
}

func (node *InitNode) buttons() [][]Button {
	return [][]Button{{
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
}

func (node *InitNode) Back(user *User, ask *Ask, vk *VK, prev StateNode) error {
	return vk.ChangeKeyboard(user.id, CreateKeyboard(node, node.buttons()))
}

func (node *InitNode) Entry(user *User, ask *Ask, vk *VK) error {
	_, err := vk.SendMessage(user.id, "Здравствуйте!", CreateKeyboard(node, node.buttons()), nil)
	return err
}

func (node *InitNode) NewMessage(user *User, ask *Ask, vk *VK, message string) (StateNode, bool, error) {
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
	}

	return nil, false, nil
}
