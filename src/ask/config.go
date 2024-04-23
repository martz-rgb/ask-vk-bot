package ask

import (
	"errors"
	"os"
	"strconv"
	"time"

	str2duration "github.com/xhit/go-str2duration/v2"
	"go.uber.org/zap"
)

type OrganizationHashtags struct {
	PollHashtag       string `json:"ASK_POLL_HASHTAG"`
	AcceptanceHashtag string `json:"ASK_ACCEPTANCE_HASHTAG"`
	FreeAnswerHashtag string `json:"ASK_FREE_ANSWER_HASHTAG"`
	LeavingHashtag    string `json:"ASK_LEAVING_HASHTAG"`
}

type Config struct {
	Timezone             int           `json:"ASK_TIMEZONE"`
	Deadline             time.Duration `json:"ASK_DEADLINE"`
	ReservationDuration  time.Duration `json:"ASK_RESERVATION_DURATION"`
	NoConfirmReservation bool          `json:"ASK_NO_CONFIRM_RESERVATION"`

	OrganizationHashtags
}

func ConfigFromEnv() *Config {
	timezone, _ := strconv.Atoi(os.Getenv("ASK_TIMEZONE"))

	deadline, err := str2duration.ParseDuration(os.Getenv("ASK_DEADLINE"))
	if err != nil {
		zap.S().Warnw("failed to parse deadline duration",
			"error", err,
			"duration", os.Getenv("ASK_DEADLINE"))
	}

	reservation, err := str2duration.ParseDuration(os.Getenv("ASK_RESERVATION_DURATION"))
	if err != nil {
		zap.S().Warnw("failed to parse reservation duration",
			"error", err,
			"reservation duration", os.Getenv("ASK_RESERVATION_DURATION"))
	}

	no_confirm_reservation, err := strconv.ParseBool(os.Getenv("ASK_NO_CONFIRM_RESERVATION"))
	if err != nil {
		zap.S().Warnw("failed to parse no confirm reservation",
			"error", err,
			"reservation duration", os.Getenv("ASK_NO_CONFIRM_RESERVATION"))
	}

	return &Config{
		Timezone:             timezone,
		Deadline:             deadline,
		ReservationDuration:  reservation,
		NoConfirmReservation: no_confirm_reservation,

		// hashtags
		OrganizationHashtags: OrganizationHashtags{
			PollHashtag:       os.Getenv("ASK_POLL_HASHTAG"),
			AcceptanceHashtag: os.Getenv("ASK_ACCEPTANCE_HASHTAG"),
			FreeAnswerHashtag: os.Getenv("ASK_FREE_ANSWER_HASHTAG"),
			LeavingHashtag:    os.Getenv("ASK_LEAVING_HASHTAG"),
		},
	}
}

func (c *Config) Validate() error {
	// timezone default is zero

	if c.Deadline == 0 {
		return errors.New("ask deadline is not provided")
	}
	if c.ReservationDuration == 0 {
		return errors.New("ask reservation duration is not provided")
	}

	// no confirm reservation default is false

	if len(c.PollHashtag) == 0 {
		return errors.New("ask poll hashtag is not provided")
	}
	if len(c.AcceptanceHashtag) == 0 {
		return errors.New("ask acceptance hashtag is not provided")
	}
	// TO-DO: free answers are additional feature
	if len(c.FreeAnswerHashtag) == 0 {
		return errors.New("ask free answer hashtag is not provided")
	}
	if len(c.LeavingHashtag) == 0 {
		return errors.New("ask leaving hashtag is not provided")
	}

	return nil
}

func (a *Ask) OrganizationHashtags() *OrganizationHashtags {
	return &a.config.OrganizationHashtags
}
