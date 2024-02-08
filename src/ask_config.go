package main

import (
	"errors"
	"os"
	"strconv"
	"time"

	str2duration "github.com/xhit/go-str2duration/v2"
	"go.uber.org/zap"
)

type AskConfig struct {
	Timezone            int           `json:"ASK_TIMEZONE"`
	Deadline            time.Duration `json:"ASK_DEADLINE"`
	ReservationDuration time.Duration `json:"ASK_RESERVATION_DURATION"`
}

func AskConfigFromEnv() *AskConfig {
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

	return &AskConfig{
		Timezone:            timezone,
		Deadline:            deadline,
		ReservationDuration: reservation,
	}
}

func (c *AskConfig) Validate() error {
	// timezone default is zero

	if c.Deadline == 0 {
		return errors.New("ask deadline is not provided")
	}
	if c.ReservationDuration == 0 {
		return errors.New("ask reservation duration is not provided")
	}

	return nil
}
