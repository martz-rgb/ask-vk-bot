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

func (node *DeadlineNode) Entry(user *User, ask *Ask, vk *VK, params Params) error {
	members, err := ask.MembersById(user.id)
	if err != nil {
		return err
	}

	deadlines := []string{}

	for _, member := range members {
		deadline, err := ask.Deadline(member.Id)
		if err != nil {
			return err
		}

		message := fmt.Sprintf("Ваш дедлайн за роль %s: %d %s %d",
			member.Role,
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

	_, err = vk.SendMessage(user.id, strings.Join(deadlines, "\n"), CreateKeyboard(node, buttons), nil)
	return nil
}

func (node *DeadlineNode) NewMessage(user *User, ask *Ask, vk *VK, message string) (StateNode, error) {
	return nil, nil
}

func (node *DeadlineNode) KeyboardEvent(user *User, ask *Ask, vk *VK, payload *CallbackPayload) (StateNode, error) {
	switch payload.Command {
	case "history":
		history, err := ask.HistoryDeadline(user.id)
		if err != nil {
			return nil, err
		}

		message, attachment, err := node.PrepareHistory(user.id, ask, vk, history)

		_, err = vk.SendMessage(user.id, message, "", api.Params{"attachment": attachment})
		return nil, err
	case "back":
		return &InitNode{}, nil
	}

	return nil, nil
}

func (node *DeadlineNode) PrepareHistory(user_id int, ask *Ask, vk *VK, history []Deadline) (message string, attachment string, err error) {
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

	id, err := vk.UploadDocument(user_id, name, bytes.NewReader([]byte(full_history)))
	if err != nil {
		return "", "", err
	}
	if id == 0 {
		return "", "", errors.New("no doc id")
	}

	attachment = fmt.Sprintf("%s%d_%d", "doc", user_id, id)

	return message, attachment, nil
}
