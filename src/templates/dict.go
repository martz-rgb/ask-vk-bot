package templates

import "ask-bot/src/ask"

type MsgGreetingData struct{}
type MsgPointsData struct{ Points int }
type MsgPointsNoHistoryData struct{}
type MsgPointsEventData struct {
    Diff  int
    Date  string
    Cause string
}
type MsgPointsShortHistoryData struct {
    Events string
    Count  int
}
type MsgReservationNewData struct{}
type MsgReservationNewConfirmationData struct{ ask.Role }
type MsgReservationNewIntroData struct{}
type MsgReservationNewSuccessData struct{ ask.Role }
type MsgReservationCancelData struct{ ask.Reservation }
type MsgReservationCancelSuccessData struct{}
type MsgReservationGreetingRequestData struct{}
type MsgReservationUnderConsiderationData struct{ ask.Reservation }
type MsgReservationInProgressData struct{ ask.Reservation }
type MsgReservationDoneData struct{ ask.Reservation }
type MsgReservationPollData struct {
    ask.Reservation
    Link string
}

var Templates = map[TemplateID]Template{MsgGreeting: {Type: (*MsgGreetingData)(nil)}, MsgPoints: {Type: (*MsgPointsData)(nil)}, MsgPointsNoHistory: {Type: (*MsgPointsNoHistoryData)(nil)}, MsgPointsEvent: {Type: (*MsgPointsEventData)(nil)}, MsgPointsShortHistory: {Type: (*MsgPointsShortHistoryData)(nil)}, MsgReservationNew: {Type: (*MsgReservationNewData)(nil)}, MsgReservationNewConfirmation: {Type: (*MsgReservationNewConfirmationData)(nil)}, MsgReservationNewIntro: {Type: (*MsgReservationNewIntroData)(nil)}, MsgReservationNewSuccess: {Type: (*MsgReservationNewSuccessData)(nil)}, MsgReservationCancel: {Type: (*MsgReservationCancelData)(nil)}, MsgReservationCancelSuccess: {Type: (*MsgReservationCancelSuccessData)(nil)}, MsgReservationGreetingRequest: {Type: (*MsgReservationGreetingRequestData)(nil)}, MsgReservationUnderConsideration: {Type: (*MsgReservationUnderConsiderationData)(nil)}, MsgReservationInProgress: {Type: (*MsgReservationInProgressData)(nil)}, MsgReservationDone: {Type: (*MsgReservationDoneData)(nil)}, MsgReservationPoll: {Type: (*MsgReservationPollData)(nil)}}
