package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/signal"

	"go.uber.org/zap"
)

type Config struct {
	GroupToken string `json:"GROUP_TOKEN"`
	AdminToken string `json:"ADMIN_TOKEN"`
	DB         string `json:"DB"`
	Schema     string `json:"SCHEMA"`
	LogFile    string `json:"LOG_FILE"`

	SecretGroupToken string `json:"SECRET_GROUP_TOKEN"`
	SecretAdminToken string `json:"SECRET_ADMIN_TOKEN"`
}

func main() {
	config := Config{
		SecretGroupToken: os.Getenv("SECRET_GROUP_TOKEN"),
		SecretAdminToken: os.Getenv("SECRET_ADMIN_TOKEN"),
		DB:               os.Getenv("DB"),
		Schema:           os.Getenv("SCHEMA"),
		LogFile:          os.Getenv("LOG_FILE"),
	}

	if len(config.SecretGroupToken) != 0 {
		file, err := os.Open(config.SecretGroupToken)
		if err != nil {
			panic(err)
		}

		token, err := io.ReadAll(file)
		if err != nil {
			panic(err)
		}
		config.GroupToken = string(token)
	}

	if len(config.SecretAdminToken) != 0 {
		file, err := os.Open(config.SecretAdminToken)
		if err != nil {
			panic(err)
		}

		token, err := io.ReadAll(file)
		if err != nil {
			panic(err)
		}
		config.AdminToken = string(token)
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

	if len(config.GroupToken) == 0 {
		panic("no group token is provided")
	}
	if len(config.AdminToken) == 0 {
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

	api, err := NewVkApi(config.GroupToken, config.AdminToken)
	if err != nil {
		panic(err)
	}

	chat_bot := NewChatBot(&InitNode{}, api, db)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	chat_bot.RunLongPoll(ctx)
}
