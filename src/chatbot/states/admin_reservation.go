package states

import (
	"ask-bot/src/ask"
	"ask-bot/src/dict"
	"ask-bot/src/form"
	"ask-bot/src/form/check"
	"ask-bot/src/paginator"
	ts "ask-bot/src/templates"
	"ask-bot/src/vk"
	"slices"
	"strconv"
)

type AdminReservation struct {
	paginator *paginator.Paginator[form.Option]
}

func (state *AdminReservation) ID() string {
	return "admin_reservation"
}

func (state *AdminReservation) options(num_reservations int, under_consideration bool) (options []form.Option) {
	if num_reservations == 0 {
		return nil
	}

	if under_consideration {
		options = append(options, form.Option{
			ID:    "considerate",
			Label: "Рассмотреть",
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
func (state *AdminReservation) Entry(user *User, c *Controls) error {
	reservations, err := c.Ask.Reservations()
	if err != nil {
		return err
	}

	message, err := ts.ParseTemplate(
		ts.MsgAdminReservations,
		ts.MsgAdminReservationsData{
			Reservations: reservations,
		},
	)
	if err != nil {
		return err
	}

	under_consideration := slices.ContainsFunc(reservations, func(r ask.Reservation) bool {
		return r.Status == ask.ReservationStatuses.UnderConsideration
	})

	config := &paginator.Config[form.Option]{
		Command: "options",

		ToLabel: form.OptionToLabel,
		ToColor: form.OptionToColor,
		ToValue: form.OptionToValue,
	}

	state.paginator = paginator.New(state.options(len(reservations), under_consideration),
		config.MustBuild())

	_, err = c.Vk.SendMessage(user.Id,
		message,
		vk.CreateKeyboard(state.ID(), state.paginator.Buttons()),
		nil)

	return err
}

func (state *AdminReservation) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	return nil, nil
}

func (state *AdminReservation) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "options":
		option, err := state.paginator.Object(payload.Value)
		if err != nil {
			return nil, err
		}

		switch option.ID {
		case "considerate":
			reservations, err := c.Ask.UnderConsiderationReservations()
			if err != nil {
				return nil, err
			}

			var options []form.Option
			for _, r := range reservations {
				options = append(options, form.Option{
					ID:    strconv.Itoa(r.VkID),
					Label: r.ShownName,
					Value: &r,
				})
			}

			reservation := form.Field{
				Name:           "reservation",
				BuildRequest:   form.AlwaysRequest(&vk.MessageParams{Text: "Выберить бронь для рассмотрения."}, options),
				ExtrudeMessage: nil,
				Check:          check.NotEmpty,
			}

			decision := form.Field{
				Name: "decision",
				BuildRequest: func(d dict.Dictionary) (*form.Request, bool, error) {
					r, err := dict.ExtractValue[*ask.Reservation](d, "reservation")
					if err != nil {
						return nil, false, err
					}

					message, err := ts.ParseTemplate(
						ts.MsgAdminReservationConsiderate,
						ts.MsgAdminReservationConsiderateData{
							Reservation: *r,
						},
					)
					if err != nil {
						return nil, false, err
					}

					forward, err := vk.ForwardParam(
						r.VkID,
						[]int{r.Introduction})
					if err != nil {
						return nil, false, err
					}

					return &form.Request{
						Message: &vk.MessageParams{
							Text:   message,
							Params: forward,
						},
						Options: []form.Option{
							{
								ID:    "confirm",
								Color: vk.PrimaryColor,
								Label: "Подтвердить",
								Value: true,
							},
							{
								ID:    "decline",
								Color: vk.SecondaryColor,
								Label: "Отклонить",
								Value: false,
							},
						},
					}, false, nil
				},
				ExtrudeMessage: nil,
				Check:          check.NotEmptyBool,
			}

			form, err := NewForm("considerate", reservation, decision)
			return NewActionNext(form), err

		case "delete":
			reservations, err := c.Ask.Reservations()
			if err != nil {
				return nil, err
			}

			var options []form.Option
			for _, r := range reservations {
				options = append(options, form.Option{
					ID:    strconv.Itoa(r.VkID),
					Label: r.ShownName,
					Value: &r,
				})
			}

			field := form.Field{
				Name: "reservation",
				BuildRequest: form.AlwaysRequest(
					&vk.MessageParams{Text: "Выберить бронь для удаления."},
					options),
				ExtrudeMessage: nil,
				Check:          check.NotEmpty,
			}

			form, err := NewForm("delete", field)
			return NewActionNext(form), err
		}
	case "paginator":
		back := state.paginator.Control(payload.Value)

		if back {
			return NewActionExit(nil), nil
		}

		return nil, c.Vk.ChangeKeyboard(user.Id,
			vk.CreateKeyboard(state.ID(), state.paginator.Buttons()))
	}

	return nil, nil
}

func (state *AdminReservation) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	if info == nil {
		return nil, state.Entry(user, c)
	}

	switch info.Payload {
	case "considerate":
		reservation, err := dict.ExtractValue[*ask.Reservation](info.Values, "reservation")
		if err != nil {
			return nil, err
		}
		decision, err := dict.ExtractValue[bool](info.Values, "decision")
		if err != nil {
			return nil, err
		}

		if decision {
			deadline, err := c.Ask.ConfirmReservation(reservation.VkID)
			if err != nil {
				return nil, err
			}

			reservation.Deadline.Time = deadline
		} else {
			err := c.Ask.DeleteReservation(reservation.VkID)
			if err != nil {
				return nil, err
			}
		}

		message, err := ts.ParseTemplate(
			ts.MsgAdminReservationConsiderated,
			ts.MsgAdminReservationConsideratedData{
				Decision:    decision,
				Reservation: *reservation,
			},
		)
		notification, err := ts.ParseTemplate(
			ts.MsgAdminReservationConsideratedNotify,
			ts.MsgAdminReservationConsideratedNotifyData{
				Decision:    decision,
				Reservation: *reservation,
			},
		)

		// notify user
		c.Notify <- &vk.MessageParams{
			Id:   reservation.VkID,
			Text: notification,
		}

		_, err = c.Vk.SendMessage(user.Id, message, "", nil)
		if err != nil {
			return nil, err
		}

	case "delete":
		reservation, err := dict.ExtractValue[*ask.Reservation](info.Values, "reservation")
		if err != nil {
			return nil, err
		}

		err = c.Ask.DeleteReservation(reservation.VkID)
		if err != nil {
			return nil, err
		}

		message, err := ts.ParseTemplate(
			ts.MsgAdminReservationDeleted,
			ts.MsgAdminReservationDeletedData{
				Reservation: *reservation,
			},
		)
		if err != nil {
			return nil, err
		}

		_, err = c.Vk.SendMessage(user.Id, message, "", nil)
		if err != nil {
			return nil, err
		}
	}

	return nil, state.Entry(user, c)
}
