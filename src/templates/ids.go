package templates

import (
	"text/template"
)

//go:generate enumutil -const=TemplateID -dict=Template -name=Templates -output=dict.go -suffix=Data -json=C:\\Users\\noise\\Documents\\ask-vk-bot\\templates.json ids.go

type TemplateID string

type Template struct {
	Templates []*template.Template
	Type      interface{} `enumutil:"type"`
}

const (
	MsgGreeting TemplateID = "msg_greeting"

	MsgPoints             TemplateID = "msg_points"
	MsgPointsNoHistory    TemplateID = "msg_points_no_history"
	MsgPointsEvent        TemplateID = "msg_points_event"
	MsgPointsShortHistory TemplateID = "msg_points_short_history"

	MsgReservationNew             TemplateID = "msg_reservation_new"
	MsgReservationNewConfirmation TemplateID = "msg_reservation_new_confirmation"
	MsgReservationNewIntro        TemplateID = "msg_reservation_new_intro"
	MsgReservationNewSuccess      TemplateID = "msg_reservation_new_success"

	MsgReservationCancel          TemplateID = "msg_reservation_cancel"
	MsgReservationCancelSuccess   TemplateID = "msg_reservation_cancel_success"
	MsgReservationGreetingRequest TemplateID = "msg_reservation_greeting_request"

	MsgReservationUnderConsideration TemplateID = "msg_reservation_under_consideration"
	MsgReservationInProgress         TemplateID = "msg_reservation_in_progress"
	MsgReservationDone               TemplateID = "msg_reservation_done"
	MsgReservationPoll               TemplateID = "msg_reservation_poll"

	MsgMemberDeadline TemplateID = "msg_member_deadline"

	MsgAdminRoles                         TemplateID = "msg_admin_roles"
	MsgAdminRolesItem                     TemplateID = "msg_admin_roles_item"
	MsgAdminReservations                  TemplateID = "msg_admin_reservations"
	MsgAdminReservationConsiderated       TemplateID = "msg_admin_reservation_considerated"
	MsgAdminReservationConsideratedNotify TemplateID = "msg_admin_reservation_considerated_notify"
	MsgAdminReservationDeleted            TemplateID = "msg_admin_reservation_deleted"
)
