package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/awnumar/memguard"
	"go.uber.org/zap"
)

type Config struct {
	SecretGroupToken string        `json:"SECRET_GROUP_TOKEN"`
	SecretAdminToken string        `json:"SECRET_ADMIN_TOKEN"`
	DB               string        `json:"DB"`
	Schema           string        `json:"SCHEMA"`
	Timeout          time.Duration `json:"TIMEOUT"`
	LogDir           string        `json:"LOG_DIR"`
}

func main() {
	config := Config{
		SecretGroupToken: os.Getenv("SECRET_GROUP_TOKEN"),
		SecretAdminToken: os.Getenv("SECRET_ADMIN_TOKEN"),
		DB:               os.Getenv("DB"),
		Schema:           os.Getenv("SCHEMA"),
		LogDir:           os.Getenv("LOG_DIR"),
	}

	config_file, err := os.Open("../config.json")
	if err == nil {
		content, err := io.ReadAll(config_file)
		config_file.Close()
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(content, &config)
		if err != nil {
			panic(err)
		}
	}

	if len(config.SecretGroupToken) == 0 {
		panic("no place of group token is provided")
	}
	if len(config.SecretAdminToken) == 0 {
		panic("no place of admin token is provided")
	}

	var group_token, admin_token *memguard.LockedBuffer

	group_file, err := os.Open(config.SecretGroupToken)
	if err != nil {
		panic(err)
	}
	group_token, err = memguard.NewBufferFromEntireReader(group_file)
	group_file.Close()
	if err != nil {
		panic(err)
	}

	admin_file, err := os.Open(config.SecretAdminToken)
	if err != nil {
		panic(err)
	}
	admin_token, err = memguard.NewBufferFromEntireReader(admin_file)
	admin_file.Close()
	if err != nil {
		panic(err)
	}

	if group_token.Size() == 0 {
		panic("group token is not provided")
	}
	if admin_token.Size() == 0 {
		panic("admin token is not provided")
	}
	if len(config.DB) == 0 {
		panic("database url is not provided")
	}
	if len(config.Schema) == 0 {
		panic("database schema is not provided")
	}
	if config.Timeout == 0 {
		config.Timeout = 1 * time.Hour
	}
	if len(config.LogDir) == 0 {
		panic("log directory is not provided")
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
		panic(err)
	}

	if err = db.Init(config.Schema); err != nil {
		panic(err)
	}

	ask := NewAsk(nil, db)

	group_api, err := NewVK(group_token)
	if err != nil {
		panic(err)
	}
	admin_api, err := NewVK(admin_token)
	if err != nil {
		panic(err)
	}
	group_token.Destroy()
	admin_token.Destroy()

	chat_bot := NewChatBot(ask, &InitNode{}, config.Timeout, group_api)
	listener := NewListener(ask, group_api, admin_api)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	chat_bot.RunLongPoll(ctx)
	listener.RunLongPoll(ctx)
}
