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

type Deadline struct{}

func (state *Deadline) ID() string {
	return "deadline"
}

func (state *Deadline) Entry(user *User, c *Controls) error {
	members, err := c.Ask.MembersByVkID(user.Id)
	if err != nil {
		return err
	}

	deadlines := []string{}

	for i := range members {
		deadline, err := c.Ask.Deadline(members[i].Id)
		if err != nil {
			return err
		}

		message := fmt.Sprintf("Ваш дедлайн за роль: %d %s %d",
			deadline.Day(),
			russian.MonthGenitive(deadline.Month()),
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

	_, err = c.Vk.SendMessage(user.Id,
		strings.Join(deadlines, "\n"),
		vk.CreateKeyboard(state.ID(), buttons),
		nil)
	return err
}

func (state *Deadline) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	return nil, nil
}

func (state *Deadline) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "history":
		history, err := c.Ask.HistoryDeadline(user.Id)
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

func (state *Deadline) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	return nil, state.Entry(user, c)
}

func (state *Deadline) PrepareHistory(user_id int, c *Controls, history []ask.Deadline) (message string, attachment string, err error) {
	if len(history) == 0 {
		return "Нет событий.", "", nil
	}

	points_noun := russian.PluralNoun("секунда", "секунды", "секунд")

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
