package main

type InitNode struct{}

func (node *InitNode) ID() string {
	return "init"
}

func (node *InitNode) Entry(user *User, ask *Ask, vk *VK, params Params) error {
	buttons := [][]Button{{
		{
			Label:   "Список ролей",
			Color:   "secondary",
			Command: (&RolesNode{}).ID(),
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

	silent, ok := params.Bool("silent")

	if ok && silent {
		return vk.ChangeKeyboard(user.id, CreateKeyboard(node, buttons))
	}

	_, err := vk.SendMessage(user.id, "Здравствуйте!", CreateKeyboard(node, buttons), nil)
	return err
}

func (node *InitNode) NewMessage(user *User, ask *Ask, vk *VK, message string) (StateNode, error) {
	return nil, nil
}

func (node *InitNode) KeyboardEvent(user *User, ask *Ask, vk *VK, payload *CallbackPayload) (StateNode, error) {
	switch payload.Command {
	case (&FAQNode{}).ID():
		return &FAQNode{}, nil
	case (&RolesNode{}).ID():
		return &RolesNode{}, nil
	case (&PointsNode{}).ID():
		return &PointsNode{}, nil
	case (&DeadlineNode{}).ID():
		return &DeadlineNode{}, nil
	}

	return nil, nil
}
