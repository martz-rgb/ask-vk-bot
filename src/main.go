package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/signal"

	"github.com/awnumar/memguard"
	"go.uber.org/zap"
)

type Config struct {
	SecretGroupToken string `json:"SECRET_GROUP_TOKEN"`
	SecretAdminToken string `json:"SECRET_ADMIN_TOKEN"`
	DB               string `json:"DB"`
	Schema           string `json:"SCHEMA"`
	LogFile          string `json:"LOG_FILE"`
}

func main() {
	config := Config{
		SecretGroupToken: os.Getenv("SECRET_GROUP_TOKEN"),
		SecretAdminToken: os.Getenv("SECRET_ADMIN_TOKEN"),
		DB:               os.Getenv("DB"),
		Schema:           os.Getenv("SCHEMA"),
		LogFile:          os.Getenv("LOG_FILE"),
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
		panic("no group token is provided")
	}
	if admin_token.Size() == 0 {
		panic("no admin token is provided")
	}
	if len(config.DB) == 0 {
		panic("no database url is provided")
	}
	if len(config.Schema) == 0 {
		panic("no database schema is provided")
	}
	if len(config.LogFile) == 0 {
		panic("no log file is provided")
	}

	log_cfg := zap.NewDevelopmentConfig()
	log_cfg.OutputPaths = []string{config.LogFile}

	logger := zap.Must(log_cfg.Build())
	defer logger.Sync()

	// use as global variable
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	db, err := NewDb(config.DB)
	if err != nil {
		panic(err)
	}

	if err = db.Init(config.Schema); err != nil {
		panic(err)
	}

	api, err := NewVkApi(group_token, admin_token)
	if err != nil {
		panic(err)
	}
	group_token.Destroy()
	admin_token.Destroy()

	chat_bot := NewChatBot(&InitNode{}, api, db)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	chat_bot.RunLongPoll(ctx)
}
