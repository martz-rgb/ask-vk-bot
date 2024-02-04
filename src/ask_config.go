package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
	"time"

	str2duration "github.com/xhit/go-str2duration/v2"
	"go.uber.org/zap"
)

type AskConfig struct {
	Timezone int           `json:"ASK_TIMEZONE"`
	Deadline time.Duration `json:"ASK_DEADLINE"`
}

func AskConfigFromEnv() *AskConfig {
	timezone, _ := strconv.Atoi(os.Getenv("ASK_TIMEZONE"))
	deadline, err := str2duration.ParseDuration(os.Getenv("ASK_DEADLINE"))
	if err != nil {
		zap.S().Warnw("failed to parse deadline duration",
			"error", err,
			"duration", os.Getenv("ASK_DEADLINE"))
	}

	return &AskConfig{
		Timezone: timezone,
		Deadline: deadline,
	}
}

func AskConfigFromFile(name string) (*AskConfig, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	c := &AskConfig{}
	err = json.Unmarshal(content, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *AskConfig) Validate() error {
	// timezone default is zero

	if c.Deadline == 0 {
		return errors.New("ask deadline is not provided")
	}

	return nil
}
