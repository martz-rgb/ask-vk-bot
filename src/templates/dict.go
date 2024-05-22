package templates

import "ask-bot/src/ask"

type MessageGreetingData struct{}
type MessagePointsData struct{ Points int }
type MessagePointsNoHistoryData struct{}
type MessagePointsEventData struct {
	Diff  int
	Date  string
	Cause string
}
type MessagePointsShortHistoryData struct {
	Events string
	Count  int
}
type MessageReservationUnderConsiderationData struct{ ask.Reservation }
type MessageReservationInProgressData struct{ ask.Reservation }
type MessageReservationDoneData struct{ ask.Reservation }
type MessageReservationPollData struct {
	ask.Reservation
	Link string
}

var Templates = map[TemplateID]Template{MessageGreeting: {Type: (*MessageGreetingData)(nil)}, MessagePoints: {Type: (*MessagePointsData)(nil)}, MessagePointsNoHistory: {Type: (*MessagePointsNoHistoryData)(nil)}, MessagePointsEvent: {Type: (*MessagePointsEventData)(nil)}, MessagePointsShortHistory: {Type: (*MessagePointsShortHistoryData)(nil)}, MessageReservationUnderConsideration: {Type: (*MessageReservationUnderConsiderationData)(nil)}, MessageReservationInProgress: {Type: (*MessageReservationInProgressData)(nil)}, MessageReservationDone: {Type: (*MessageReservationDoneData)(nil)}, MessageReservationPoll: {Type: (*MessageReservationPollData)(nil)}}
