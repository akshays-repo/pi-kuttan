package main

import (
	"log"

	"mypibot-go/internal/bot"
	"mypibot-go/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize and start bot
	b, err := bot.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Start the bot
	b.Start()
}
