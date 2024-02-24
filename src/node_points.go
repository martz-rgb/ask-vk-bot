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

const MaxLengthHistory int = 10

type PointsNode struct {
}

func (node *PointsNode) ID() string {
	return "points"
}

func (node *PointsNode) Entry(user *User, c *Controls) error {
	points, err := c.Ask.Points(user.id)
	if err != nil {
		return err
	}

	buttons := [][]vk.Button{
		{{
			Label: "Потратить",
			Color: vk.SecondaryColor,

			Command: "spend",
		}},
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

	message := fmt.Sprintf("Ваше текущее количество баллов: %d", points)

	_, err = c.Vk.SendMessage(user.id, message, vk.CreateKeyboard(node.ID(), buttons), nil)
	return err
}

func (node *PointsNode) NewMessage(user *User, c *Controls, message *vk.Message) (StateNode, bool, error) {
	return nil, false, nil
}

func (node *PointsNode) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (StateNode, bool, error) {
	switch payload.Command {
	case "spend":
		message := `Пока что не на что тратить баллы.`
		_, err := c.Vk.SendMessage(user.id, message, "", nil)
		return nil, false, err
	case "history":
		history, err := c.Ask.HistoryPoints(user.id)
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

func (node *PointsNode) Back(user *User, c *Controls, prev_state StateNode) (bool, error) {
	return false, nil
}

func (node *PointsNode) PrepareHistory(user_id int, c *Controls, history []ask.Points) (message string, attachment string, err error) {
	if len(history) == 0 {
		return "Вы еще не получали баллы в нашем сообществе.", "", nil
	}

	points_noun := PluralNoun("балл", "балла", "баллов")

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
