package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
)

const MaxLengthHistory int = 10

type KeyboardAction func(int, *Ask, *VK, *CallbackPayload) StateNode

type PointsNode struct {
}

func (node *PointsNode) ID() string {
	return "points"
}

func (node *PointsNode) Entry(user *User, ask *Ask, vk *VK, params Params) error {
	points, err := ask.Points(user.id)
	if err != nil {
		return err
	}

	buttons := [][]Button{
		{{
			Label: "Потратить",
			Color: SecondaryColor,

			Command: "spend",
		}},
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

	message := fmt.Sprintf("Ваше текущее количество баллов: %d", points)

	_, err = vk.SendMessage(user.id, message, CreateKeyboard(node, buttons), nil)
	return err
}

func (node *PointsNode) NewMessage(user *User, ask *Ask, vk *VK, message string) (StateNode, error) {
	return nil, nil
}

func (node *PointsNode) KeyboardEvent(user *User, ask *Ask, vk *VK, payload *CallbackPayload) (StateNode, error) {
	switch payload.Command {
	case "spend":
		message := `Пока что не на что тратить баллы.`
		_, err := vk.SendMessage(user.id, message, "", nil)
		return nil, err
	case "history":
		history, err := ask.HistoryPoints(user.id)
		if err != nil {
			return nil, err
		}

		message, attachment, err := node.PrepareHistory(user.id, ask, vk, history)
		if err != nil {
			return nil, err
		}

		_, err = vk.SendMessage(user.id, message, "", api.Params{"attachment": attachment})
		return nil, err
	case "back":
		return &InitNode{}, nil
	}

	return nil, nil
}

func (node *PointsNode) PrepareHistory(user_id int, ask *Ask, vk *VK, history []Points) (message string, attachment string, err error) {
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
