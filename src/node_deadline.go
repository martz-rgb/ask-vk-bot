package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
)

type DeadlineNode struct{}

func (node *DeadlineNode) ID() string {
	return "deadline"
}

func (node *DeadlineNode) Entry(user *User, c *Controls) error {
	members, roles, err := user.MembersRoles(c.Ask)
	if err != nil {
		return err
	}

	deadlines := []string{}

	for i := range members {
		deadline, err := c.Ask.Deadline(members[i].Id)
		if err != nil {
			return err
		}

		message := fmt.Sprintf("Ваш дедлайн за роль %s: %d %s %d",
			roles[i].ShownName,
			deadline.Day(),
			MonthGenitive(deadline.Month()),
			deadline.Year())
		deadlines = append(deadlines, message)
	}

	buttons := [][]Button{
		{
			{
				Label: "История",
				Color: SecondaryColor,

				Command: "history",
			},
			{
				Label: "Назад",
				Color: NegativeColor,

				Command: "back",
			},
		},
	}

	_, err = c.Vk.SendMessage(user.id, strings.Join(deadlines, "\n"), CreateKeyboard(node, buttons), nil)
	return err
}

func (node *DeadlineNode) NewMessage(user *User, c *Controls, message *Message) (StateNode, bool, error) {
	return nil, false, nil
}

func (node *DeadlineNode) KeyboardEvent(user *User, c *Controls, payload *CallbackPayload) (StateNode, bool, error) {
	switch payload.Command {
	case "history":
		history, err := c.Ask.HistoryDeadline(user.id)
		if err != nil {
			return nil, false, err
		}

		message, attachment, err := node.PrepareHistory(user.id, c, history)
		if err != nil {
			return nil, false, err
		}

		_, err = c.Vk.SendMessage(user.id, message, "", api.Params{"attachment": attachment})
		return nil, false, err
	case "back":
		return nil, true, nil
	}

	return nil, false, nil
}

func (node *DeadlineNode) Back(user *User, c *Controls, prev_state StateNode) (bool, error) {
	return false, nil
}

func (node *DeadlineNode) PrepareHistory(user_id int, c *Controls, history []Deadline) (message string, attachment string, err error) {
	if len(history) == 0 {
		return "Нет событий.", "", nil
	}

	points_noun := PluralNoun("секунда", "секунды", "секунд")

	events := []string{}
	for _, event := range history {
		sign := "+"
		sign_word := "получили"
		if event.Diff < 0 {
			sign = "−"
			sign_word = "потеряли"
			event.Diff = -event.Diff
		}

		events = append(events, fmt.Sprintf(
			"%s Вы %s %d %s %d %s %d в %s.\n   Причина: \"%s\".\n",
			sign,
			sign_word,
			event.Diff,
			points_noun(event.Diff),
			event.Timestamp.Day(),
			MonthGenitive(event.Timestamp.Month()),
			event.Timestamp.Year(),
			event.Timestamp.Format(time.TimeOnly),
			event.Cause))
	}

	if len(history) <= MaxLengthHistory {
		return strings.Join(events, "\n"), "", nil
	}
	record_noun := PluralNoun("запись", "записи", "записей")

	message = strings.Join(events[:MaxLengthHistory], "\n")
	message += fmt.Sprintf("\n... и еще %d %s. Смотрите полную историю в прикрепленном файле.",
		len(history)-MaxLengthHistory,
		record_noun(len(history)-MaxLengthHistory))

	name := fmt.Sprintf("full_history_%d_%s.txt", user_id, time.Now().Format(time.DateOnly))
	full_history := strings.Join(events, "\n")

	id, err := c.Vk.UploadDocument(user_id, name, bytes.NewReader([]byte(full_history)))
	if err != nil {
		return "", "", err
	}
	if id == 0 {
		return "", "", errors.New("no doc id")
	}

	attachment = fmt.Sprintf("%s%d_%d", "doc", user_id, id)

	return message, attachment, nil
}
