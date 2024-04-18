package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type Config struct {
	GroupID          int           `json:"GROUP_ID"`
	SecretGroupToken string        `json:"SECRET_GROUP_TOKEN"`
	SecretAdminToken string        `json:"SECRET_ADMIN_TOKEN"`
	DB               string        `json:"DB"`
	AllowDeletion    bool          `json:"ALLOW_DELETION"`
	Schema           string        `json:"SCHEMA"`
	Timeout          time.Duration `json:"TIMEOUT"`
	LogDir           string        `json:"LOG_DIR"`
	UpdatePostponed  time.Duration `json:"UPDATE_POSTPONED"`
}

func ConfigFromEnv() *Config {
	group_id, err := strconv.Atoi(os.Getenv("GROUP_ID"))
	if err != nil {
		zap.S().Warnw("failed to parse group id",
			"error", err,
			"group id", os.Getenv("GROUP_ID"))
	}
	allow_deletion, _ := strconv.ParseBool(os.Getenv("ALLOW_DELETION"))
	timeout, err := time.ParseDuration(os.Getenv("TIMEOUT"))
	if err != nil {
		zap.S().Warnw("failed to parse timeout",
			"error", err,
			"timeout", os.Getenv("TIMEOUT"))
	}
	update, err := time.ParseDuration(os.Getenv("UPDATE_POSTPONED"))
	if err != nil {
		zap.S().Warnw("failed to parse update postponed",
			"error", err,
			"update", os.Getenv("UPDATE_POSTPONED"))
	}

	return &Config{
		GroupID:          group_id,
		SecretGroupToken: os.Getenv("SECRET_GROUP_TOKEN"),
		SecretAdminToken: os.Getenv("SECRET_ADMIN_TOKEN"),
		DB:               os.Getenv("DB"),
		AllowDeletion:    allow_deletion,
		Schema:           os.Getenv("SCHEMA"),
		LogDir:           os.Getenv("LOG_DIR"),
		Timeout:          timeout,
		UpdatePostponed:  update,
	}
}

func ConfigFromFile(name string) (*Config, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	err = json.Unmarshal(content, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) Validate() error {
	if c.GroupID == 0 {
		return errors.New("group id is not provided")
	}
	if len(c.SecretGroupToken) == 0 {
		return errors.New("place of group token is not provided")
	}
	if len(c.SecretAdminToken) == 0 {
		return errors.New("place of admin token is not provided")
	}
	if len(c.DB) == 0 {
		return errors.New("database url is not provided")
	}
	// no need to check allow deletion, default is false
	if len(c.Schema) == 0 {
		return errors.New("database schema is not provided")
	}
	if len(c.LogDir) == 0 {
		return errors.New("log directory is not provided")
	}

	if c.Timeout == 0 {
		c.Timeout = 1 * time.Hour
	}
	if c.UpdatePostponed == 0 {
		c.UpdatePostponed = 1 * time.Minute
	}

	return nil
}
