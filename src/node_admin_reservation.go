package main

import (
	"ask-bot/src/ask"
	"ask-bot/src/dict"
	"ask-bot/src/form"
	"ask-bot/src/paginator"
	"ask-bot/src/vk"
	"fmt"
	"strconv"
	"strings"
)

type AdminReservationNode struct {
	paginator *paginator.Paginator[form.Option]
}

func (node *AdminReservationNode) ID() string {
	return "admin_reservation"
}

func (node *AdminReservationNode) options(num_reservations int, under_consideration bool) (options []form.Option) {
	if num_reservations == 0 {
		return nil
	}

	if under_consideration {
		options = append(options, form.Option{
			ID:    "confirm",
			Label: "Подтвердить",
			Color: vk.PrimaryColor,
		})
	}

	options = append(options, form.Option{
		ID:    "delete",
		Label: "Удалить",
		Color: vk.SecondaryColor,
	})

	return
}

// To-DO: print all reservations
func (node *AdminReservationNode) Entry(user *User, c *Controls) error {
	reservations, err := c.Ask.ReservationsDetails()
	if err != nil {
		return err
	}

	var details []string
	var under_consideration bool
	for i, r := range reservations {
		message := fmt.Sprintf("%d. %s\n user: @id%d\n status: %s\n deadline: %s",
			i+1,
			r.Hashtag,
			r.VkID,
			r.Status,
			r.Deadline.Time)
		details = append(details, message)

		if r.Status == ask.ReservationStatuses.UnderConsideration {
			under_consideration = true
		}
	}

	var message string

	if len(details) > 0 {
		message = strings.Join(details, "\n")
	} else {
		message = "Сейчас нет броней."
	}

	config := &paginator.Config[form.Option]{
		Command: "options",

		ToLabel: form.OptionToLabel,
		ToColor: form.OptionToColor,
		ToValue: form.OptionToValue,
	}

	node.paginator = paginator.New[form.Option](node.options(len(reservations), under_consideration),
		config.MustBuild())

	_, err = c.Vk.SendMessage(user.id,
		message,
		vk.CreateKeyboard(node.ID(), node.paginator.Buttons()),
		nil)

	return err
}

func (node *AdminReservationNode) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	return nil, nil
}

func (node *AdminReservationNode) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "options":
		option, err := node.paginator.Object(payload.Value)
		if err != nil {
			return nil, err
		}

		switch option.ID {
		case "confirm":
			reservations, err := c.Ask.UnderConsiderationReservationsDetails()
			if err != nil {
				return nil, err
			}

			var options []form.Option
			for _, r := range reservations {
				options = append(options, form.Option{
					ID:    strconv.Itoa(r.Id),
					Label: r.ShownName,
					Value: &r,
				})
			}

			field := form.NewField(
				"reservation",
				&vk.MessageParams{
					Text: "Выберить бронь для рассмотрения.",
				},
				options,
				nil,
				NotEmpty,
				ConfirmReservationField,
			)

			return NewActionNext(NewFormNode("confirm", nil, field)), nil

		case "delete":
			reservations, err := c.Ask.ReservationsDetails()
			if err != nil {
				return nil, err
			}

			var options []form.Option
			for _, r := range reservations {
				options = append(options, form.Option{
					ID:    strconv.Itoa(r.Id),
					Label: r.ShownName,
					Value: &r,
				})
			}

			field := form.NewField(
				"reservation",
				&vk.MessageParams{
					Text: "Выберить бронь для удаления.",
				},
				options,
				nil,
				NotEmpty,
				nil,
			)

			return NewActionNext(NewFormNode("delete", ConfirmReservationDeletion, field)), nil
		}
	case "paginator":
		back := node.paginator.Control(payload.Value)

		if back {
			return NewActionExit(nil), nil
		}

		return nil, c.Vk.ChangeKeyboard(user.id,
			vk.CreateKeyboard(node.ID(), node.paginator.Buttons()))
	}

	return nil, nil
}

func (node *AdminReservationNode) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	if info == nil {
		return nil, node.Entry(user, c)
	}

	switch info.Payload {
	case "confirm":
		reservation, err := dict.ExtractValue[*ask.ReservationDetail](info.Values, "reservation")
		if err != nil {
			return nil, err
		}
		action, err := dict.ExtractValue[bool](info.Values, "details", "action")
		if err != nil {
			return nil, err
		}

		var message, notification_message string
		if action == true {
			deadline, err := c.Ask.ConfirmReservation(reservation.Id)
			if err != nil {
				return nil, err
			}

			message = fmt.Sprintf("Бронь на %s была успешно подтверждена.",
				reservation.AccusativeName)
			notification_message = fmt.Sprintf("Ваша бронь на %s успешно подтверждена! Вам нужно отрисовать приветствие до %s.",
				reservation.AccusativeName,
				deadline)
		} else {
			err := c.Ask.DeleteReservation(reservation.Id)
			if err != nil {
				return nil, err
			}

			message = fmt.Sprintf("Бронь на %s была успешно удалена.",
				reservation.AccusativeName)
			notification_message = fmt.Sprintf("Ваша бронь на %s, к сожалению, отклонена. Попробуйте еще раз позже!",
				reservation.AccusativeName)
		}

		// notify user
		notification := &vk.MessageParams{
			Id:   reservation.VkID,
			Text: notification_message,
		}
		c.Notify <- notification

		_, err = c.Vk.SendMessage(user.id, message, "", nil)
		if err != nil {
			return nil, err
		}

	case "delete":
		reservation, err := dict.ExtractValue[*ask.ReservationDetail](info.Values, "reservation")
		if err != nil {
			return nil, err
		}

		err = c.Ask.DeleteReservation(reservation.Id)
		if err != nil {
			return nil, err
		}

		message := fmt.Sprintf("Бронь на %s от @id%d была успешно удалена.",
			reservation.AccusativeName,
			reservation.VkID)
		_, err = c.Vk.SendMessage(user.id, message, "", nil)
		if err != nil {
			return nil, err
		}
	}

	return nil, node.Entry(user, c)
}
