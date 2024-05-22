package templates

import (
	"text/template"
)

//go:generate enumutil -const=TemplateID -dict=Template -name=Templates -output=dict.go -suffix=Data ids.go

type TemplateID string

type Template struct {
	Templates []*template.Template
	Type      interface{} `enumutil:"type"`
}

const (
	MessageGreeting TemplateID = "msg_greeting"

	MessagePoints             TemplateID = "msg_points"
	MessagePointsNoHistory    TemplateID = "msg_points_no_history"
	MessagePointsEvent        TemplateID = "msg_points_event"
	MessagePointsShortHistory TemplateID = "msg_points_short_history"

	MessageReservationUnderConsideration TemplateID = "msg_reservation_under_consideration"
	MessageReservationInProgress         TemplateID = "msg_reservation_in_progress"
	MessageReservationDone               TemplateID = "msg_reservation_done"
	MessageReservationPoll               TemplateID = "msg_reservation_poll"
)
