package main

import (
	"ask-bot/src/ask"
	"ask-bot/src/vk"
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

	buttons := [][]vk.Button{
		{
			{
				Label: "История",
				Color: vk.SecondaryColor,

				Command: "history",
			},
			{
				Label: "Назад",
				Color: vk.NegativeColor,

				Command: "back",
			},
		},
	}

	_, err = c.Vk.SendMessage(user.id,
		strings.Join(deadlines, "\n"),
		vk.CreateKeyboard(node.ID(), buttons),
		nil)
	return err
}

func (node *DeadlineNode) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	return nil, nil
}

func (node *DeadlineNode) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "history":
		history, err := c.Ask.HistoryDeadline(user.id)
		if err != nil {
			return nil, err
		}

		message, attachment, err := node.PrepareHistory(user.id, c, history)
		if err != nil {
			return nil, err
		}

		_, err = c.Vk.SendMessage(user.id, message, "", api.Params{"attachment": attachment})
		return nil, err
	case "back":
		return NewActionExit(&ExitInfo{}), nil
	}

	return nil, nil
}

func (node *DeadlineNode) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	return nil, node.Entry(user, c)
}

func (node *DeadlineNode) PrepareHistory(user_id int, c *Controls, history []ask.Deadline) (message string, attachment string, err error) {
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
