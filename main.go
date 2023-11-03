package main

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	GroupToken string `json:"GROUP_TOKEN"`
	AdminToken string `json:"ADMIN_TOKEN"`
	DataBase   string `json:"DATABASE"`

	AppId        int64  `json:"APP_ID"`
	ProtectedKey string `json:"PROTECTED_KEY"`
	ServerKey    string `json:"SERVER_KEY"`
}

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
	if len(config.DataBase) == 0 {
		panic("no database url in config file")
	}

	api, err := NewVkApi(config.GroupToken, config.AdminToken)
	if err != nil {
		panic(err)
	}

	db, err := NewDb(config.DataBase)
	if err != nil {
		panic(err)
	}

	if err = db.Init(); err != nil {
		panic(err)
	}

	//db.LoadCsv()

	chat_bot := NewChatBot(&InitNode{}, api, db)
	chat_bot.RunLongPoll()
}
