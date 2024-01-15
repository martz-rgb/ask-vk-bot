package main

import (
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
	Schema           string        `json:"SCHEMA"`
	Timeout          time.Duration `json:"TIMEOUT"`
	LogDir           string        `json:"LOG_DIR"`
}

func ConfigFromEnv() *Config {
	group_id, _ := strconv.Atoi(os.Getenv("GROUP_ID"))

	return &Config{
		GroupID:          group_id,
		SecretGroupToken: os.Getenv("SECRET_GROUP_TOKEN"),
		SecretAdminToken: os.Getenv("SECRET_ADMIN_TOKEN"),
		DB:               os.Getenv("DB"),
		Schema:           os.Getenv("SCHEMA"),
		LogDir:           os.Getenv("LOG_DIR"),
	}
}

func ConfigFromFile(name string) (*Config, error) {
	file, err := os.Open("../config.json")
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
	}

	err = config.Validate()
	if err != nil {
		log.Fatal(err)
	}

	log_cfg := zap.NewDevelopmentConfig()
	log_cfg.OutputPaths = []string{filepath.Join(config.LogDir, "chat-bot.log")}

	logger := zap.Must(log_cfg.Build())
	defer logger.Sync()

	// use as global variable
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	db, err := NewDB(config.DB)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Init(config.Schema); err != nil {
		log.Fatal(err)
	}

	ask := NewAsk(nil, db)

	group, err := NewVKFromFile(config.SecretGroupToken)
	if err != nil {
		log.Fatal(err)
	}
	admin, err := NewVKFromFile(config.SecretAdminToken)
	if err != nil {
		log.Fatal(err)
	}

	chat_bot := NewChatBot(ask, &InitNode{}, config.Timeout, config.GroupID, group)
	listener := NewListener(ask, config.GroupID, group, admin)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go listener.RunLongPoll(ctx, wg)
	go chat_bot.RunLongPoll(ctx, wg)

	fmt.Println("run", config.GroupID)

	wg.Wait()
}
