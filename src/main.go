package main

import (
	"ask-bot/src/ask"
	"ask-bot/src/chatbot"
	"ask-bot/src/events"
	"ask-bot/src/listener"
	"ask-bot/src/templates"
	"ask-bot/src/vk"
	"ask-bot/src/watcher"
	"ask-bot/src/watcher/postponed"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
)

func CreateLogger(logdir, filename string) *zap.Logger {
	log_cfg := zap.NewDevelopmentConfig()
	log_cfg.OutputPaths = []string{filepath.Join(logdir, filename)}
	log_cfg.DisableStacktrace = true

	return zap.Must(log_cfg.Build())
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

	// make main logger
	logger := CreateLogger(config.LogDir, "info.log")
	defer logger.Sync()

	// use as global variable
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	// make chatbot's, listener's, postponed's loggers
	bot_logger := CreateLogger(config.LogDir, "chatbot.log")
	listener_logger := CreateLogger(config.LogDir, "listener.log")
	watcher_logger := CreateLogger(config.LogDir, "watcher.log")

	// make templates
	err = templates.NewFromFile(config.Templates)
	if err != nil {
		zap.S().Fatalw("failed to initialize templates",
			"error", err)
	}

	// make ask layer upon db
	ask_config := ask.ConfigFromEnv()

	err = ask_config.Validate()
	if err != nil {
		zap.S().Fatalw("failed to validate ask config",
			"error", err)
	}
	a := ask.New(ask_config)

	// init db + migrate
	err = a.Init(config.DB, config.Schema, config.AllowDeletion)
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

	// linked parts init
	notify_user := make(chan *vk.MessageParams)
	postponed := postponed.New()
	notify_event := make(chan events.Event)

	w := watcher.New(&watcher.Controls{
		Ask:         a,
		Admin:       admin,
		Group:       group,
		NotifyUser:  notify_user,
		NotifyEvent: notify_event,
	},
		config.UpdatePostponed,
		postponed,
		watcher_logger.Sugar())

	c := chatbot.New(&chatbot.Controls{
		Vk:          group,
		Ask:         a,
		Notify:      notify_user,
		Postponed:   postponed,
		NotifyEvent: notify_event,
	},
		config.Timeout,
		bot_logger.Sugar())

	l := listener.New(&listener.Controls{
		Ask:        a,
		Admin:      admin,
		Group:      group,
		NotifyUser: notify_user,
	},
		listener_logger.Sugar())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	wg := &sync.WaitGroup{}
	wg.Add(3)

	go c.Run(ctx, wg)
	go l.Run(ctx, wg)
	go w.Run(ctx, wg)

	fmt.Println("run", config.GroupID)

	wg.Wait()
}
