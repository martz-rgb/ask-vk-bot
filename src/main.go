package main

import (
	"ask-bot/src/ask"
	"ask-bot/src/vk"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
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

	return &Config{
		GroupID:          group_id,
		SecretGroupToken: os.Getenv("SECRET_GROUP_TOKEN"),
		SecretAdminToken: os.Getenv("SECRET_ADMIN_TOKEN"),
		DB:               os.Getenv("DB"),
		AllowDeletion:    allow_deletion,
		Schema:           os.Getenv("SCHEMA"),
		LogDir:           os.Getenv("LOG_DIR"),
		Timeout:          timeout,
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

	return nil
}

func main() {
	config := ConfigFromEnv()

	// development purposes
	dev, err := ConfigFromFile("../config.json")
	if err == nil {
		config = dev
	} else {
		log.Printf("failed to read config from file: %s\n", err)
	}

	err = config.Validate()
	if err != nil {
		log.Fatalf("failed to validate config: %s\n", err)
	}

	// make debug log
	log_cfg := zap.NewDevelopmentConfig()
	log_cfg.OutputPaths = []string{filepath.Join(config.LogDir, "info.log")}

	logger := zap.Must(log_cfg.Build())
	defer logger.Sync()

	// use as global variable
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	// make chatbot log
	log_cfg = zap.NewDevelopmentConfig()
	log_cfg.OutputPaths = []string{filepath.Join(config.LogDir, "chatbot.log")}
	log_cfg.DisableStacktrace = true

	bot_logger := zap.Must(log_cfg.Build())

	// make ask layer upon db
	ask_config := ask.ConfigFromEnv()

	err = ask_config.Validate()
	if err != nil {
		zap.S().Fatalw("failed to validate ask config",
			"error", err)
	}
	ask := ask.New(ask_config)

	// init db + migrate
	err = ask.Init(config.DB, config.Schema, config.AllowDeletion)
	if err != nil {
		zap.S().Fatalw("failed to init ask",
			"error", err)
	}

	// vk api's init
	group, err := vk.NewFromFile(config.SecretGroupToken, config.GroupID)
	if err != nil {
		zap.S().Fatalw("failed to create group vk api from file",
			"error", err)
	}
	admin, err := vk.NewFromFile(config.SecretAdminToken, config.GroupID)
	if err != nil {
		zap.S().Fatalw("failed to create admin vk api from file",
			"error", err)
	}

	chat_bot := NewChatBot(ask, group, &InitNode{}, config.Timeout, bot_logger.Sugar())
	listener := NewListener(ask, group, admin)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go listener.RunLongPoll(ctx, wg)
	go chat_bot.RunLongPoll(ctx, wg)

	fmt.Println("run", config.GroupID)

	wg.Wait()
}
