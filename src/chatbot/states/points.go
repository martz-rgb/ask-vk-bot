package states

import (
	"ask-bot/src/ask"
	"ask-bot/src/russian"
	"ask-bot/src/templates"
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

	message, err := templates.ParseTemplate(
		templates.MessagePoints,
		templates.MessagePointsData{
			Points: points,
		},
	)
	if err != nil {
		return err
	}

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
		message, err := templates.ParseTemplate(
			templates.MessagePointsNoHistory,
			templates.MessagePointsNoHistoryData{},
		)
		return message, "", err
	}

	events := make([]string, len(history))
	for i, event := range history {
		e, err := templates.ParseTemplate(
			templates.MessagePointsEvent,
			templates.MessagePointsEventData{
				Diff: event.Diff,
				Date: fmt.Sprintf("%d %s %d",
					event.Timestamp.Day(),
					russian.MonthGenitive(event.Timestamp.Month()),
					event.Timestamp.Year()),
				Cause: event.Cause,
			},
		)
		if err != nil {
			return "", "", err
		}

		events[i] = e
	}

	if len(history) <= MaxLengthHistory {
		return strings.Join(events, "\n"), "", nil
	}

	message, err = templates.ParseTemplate(
		templates.MessagePointsShortHistory,
		templates.MessagePointsShortHistoryData{
			Events: strings.Join(events[:MaxLengthHistory], "\n"),
			Count:  len(history) - MaxLengthHistory,
		},
	)
	if err != nil {
		return "", "", nil
	}

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
