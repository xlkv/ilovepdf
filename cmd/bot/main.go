package main

import (
	"log"

	"github.com/xlkv/ilovepdf/internal/bot"
	"github.com/xlkv/ilovepdf/internal/config"
)

func main() {
	log.Println("Starting iLovePDF Telegram Bot Initialization...")
	cfg := config.Load()

	b, err := bot.NewBot(cfg)
	if err != nil {
		log.Fatalf("Fatal error creating bot: %v", err)
	}

	log.Printf("Bot connected successfully! User: @%s (ID: %d)", b.GetBotUser().Username, b.GetBotUser().Id)

	if err := b.Start(); err != nil {
		log.Fatalf("Fatal error running bot polling: %v", err)
	}
}
