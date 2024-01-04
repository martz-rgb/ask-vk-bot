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
	LogFile    string `json:"LOG_FILE"`

	AppId        int64  `json:"APP_ID"`
	ProtectedKey string `json:"PROTECTED_KEY"`
	ServerKey    string `json:"SERVER_KEY"`
}

func main() {
	config_file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}

	content, err := io.ReadAll(config_file)
	config_file.Close()
	if err != nil {
		panic(err)
	}

	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		panic(err)
	}

	if len(config.GroupToken) == 0 {
		panic("no group token in config file")
	}
	if len(config.AdminToken) == 0 {
		panic("no admin token in config file")
	}
	if len(config.DB) == 0 {
		panic("no database url in config file")
	}
	if len(config.LogFile) == 0 {
		panic("no log file in config file")
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

	if err = db.Init("schema.sql"); err != nil {
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
