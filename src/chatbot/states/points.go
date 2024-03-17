package states

import (
	"ask-bot/src/ask"
	"ask-bot/src/russian"
	"ask-bot/src/vk"
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
)

const MaxLengthHistory int = 10

type Points struct {
}

func (state *Points) ID() string {
	return "points"
}

func (state *Points) Entry(user *User, c *Controls) error {
	points, err := c.Ask.PointsByVkID(user.Id)
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

	_, err = c.Vk.SendMessage(user.Id, message, vk.CreateKeyboard(state.ID(), buttons), nil)
	return err
}

func (state *Points) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	return nil, nil
}

func (state *Points) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "spend":
		message := `Пока что не на что тратить баллы.`
		_, err := c.Vk.SendMessage(user.Id, message, "", nil)
		return nil, err
	case "history":
		history, err := c.Ask.HistoryPointsByVkID(user.Id)
		if err != nil {
			return nil, err
		}

		message, attachment, err := state.PrepareHistory(user.Id, c, history)
		if err != nil {
			return nil, err
		}

		_, err = c.Vk.SendMessage(user.Id, message, "", api.Params{"attachment": attachment})
		return nil, err
	case "back":
		return NewActionExit(nil), nil
	}

	return nil, nil
}

func (state *Points) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	return nil, state.Entry(user, c)
}

func (state *Points) PrepareHistory(user_id int, c *Controls, history []ask.Points) (message string, attachment string, err error) {
	if len(history) == 0 {
		return "Вы еще не получали баллы в нашем сообществе.", "", nil
	}

	points_noun := russian.PluralNoun("балл", "балла", "баллов")

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
			russian.MonthGenitive(event.Timestamp.Month()),
			event.Timestamp.Year(),
			event.Timestamp.Format(time.TimeOnly),
			event.Cause))
	}

	if len(history) <= MaxLengthHistory {
		return strings.Join(events, "\n"), "", nil
	}
	record_noun := russian.PluralNoun("запись", "записи", "записей")

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
