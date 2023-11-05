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

var logger *zap.SugaredLogger

func main() {
	config_file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer config_file.Close()

	content, _ := io.ReadAll(config_file)

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

	// global variable
	logger = zap.Must(log_cfg.Build()).Sugar()
	defer logger.Sync()

	api, err := NewVkApi(config.GroupToken, config.AdminToken)
	if err != nil {
		panic(err)
	}

	db, err := NewDb(config.DB)
	if err != nil {
		panic(err)
	}

	if err = db.Init("schema.sql"); err != nil {
		panic(err)
	}

	//db.LoadCsv()

	chat_bot := NewChatBot(&InitNode{}, api, db)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	chat_bot.RunLongPoll(ctx)
}
