package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type AdminReservationNode struct {
	reservation *Reservation
	paginator   *Paginator[Reservation]
}

func (node *AdminReservationNode) ID() string {
	return "admin_reservation"
}

func (node *AdminReservationNode) Entry(user *User, ask *Ask, vk *VK) error {
	reservations, err := ask.UnderConsiderationReservations()
	if err != nil {
		return err
	}

	to_label := func(r Reservation) string {
		return r.Role
	}

	to_value := func(r Reservation) string {
		return strconv.Itoa(r.Id)
	}

	node.paginator = NewPaginator[Reservation](reservations,
		"reservations",
		RowsCount,
		ColsCount,
		to_label,
		to_value)

	return vk.ChangeKeyboard(user.id,
		CreateKeyboard(node, node.paginator.Buttons()))
}

func (node *AdminReservationNode) NewMessage(user *User, ask *Ask, vk *VK, message *Message) (StateNode, bool, error) {
	return nil, false, nil
}

func (node *AdminReservationNode) KeyboardEvent(user *User, ask *Ask, vk *VK, payload *CallbackPayload) (StateNode, bool, error) {
	switch payload.Command {
	case "reservations":
		reservation, err := node.paginator.Object(payload.Value)
		if err != nil {
			return nil, false, err
		}

		node.reservation = reservation
		text := fmt.Sprintf("Роль: %s\nСтраница: https://vk.com/id%d\nДедлайн: %s\n",
			reservation.Role,
			reservation.VkID,
			reservation.Deadline)

		forward, err := json.Marshal(struct {
			PeerId      int   `json:"peer_id"`
			MessagesIds []int `json:"message_ids"`
		}{
			reservation.VkID,
			[]int{reservation.Info},
		})
		if err != nil {
			return nil, false, err
		}
		params := api.Params{
			"forward": string(forward),
		}

		request := &RequestMessage{
			Text:   text,
			Params: params,
		}

		field := NewConfirmReservation(request)
		form, err := NewForm(field)
		if err != nil {
			return nil, false, err
		}

		return &FormNode{
			Form: form,
		}, false, nil

	case "paginator":
		back := node.paginator.Control(payload.Value)

		if back {
			return nil, true, nil
		}

		return nil, false, vk.ChangeKeyboard(user.id,
			CreateKeyboard(node, node.paginator.Buttons()))
	}
	return nil, false, nil
}

func (node *AdminReservationNode) Back(user *User, ask *Ask, vk *VK, prev_state StateNode) (bool, error) {
	form, ok := prev_state.(*FormNode)
	if !ok {
		return false, nil
	}

	if !form.FilledOut {
		return false, node.Entry(user, ask, vk)
	}

	value, err := form.Form.Value(0)
	if err != nil {
		err := errors.New("form is not fullfilled")
		return false, zaperr.Wrap(err, "",
			zap.Any("form", form.Form))
	}

	confirm, err := ConvertValue[bool](value)
	if err != nil {
		return false, err
	}

	if confirm {
		err := ask.ChangeReservationStatus(node.reservation.Id,
			ReservationStatuses.InProgress)
		if err != nil {
			return false, err
		}
	} else {
		err := ask.DeleteReservation(node.reservation.Id)
		if err != nil {
			return false, err
		}
	}

	// notify user
	text := fmt.Sprintf("Статус вашей брони: %t\n", confirm)
	_, err = vk.SendMessage(node.reservation.VkID, text, "", nil)
	if err != nil {
		return false, err
	}

	return false, node.Entry(user, ask, vk)
}
