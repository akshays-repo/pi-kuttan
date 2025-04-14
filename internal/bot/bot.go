package bot

import (
	"log"

	"mypibot-go/internal/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api          *tgbotapi.BotAPI
	allowedUsers map[int64]bool
	handler      *Handler
}

func New(cfg *config.Config) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, err
	}

	// Convert allowed users to map for O(1) lookup
	allowedUsers := make(map[int64]bool)
	for _, id := range cfg.AllowedUsers {
		allowedUsers[id] = true
	}

	// Pass the bot instance to the handler
	return &Bot{
		api:          api,
		allowedUsers: allowedUsers,
		handler:      NewHandler(api),
	}, nil
}

func (b *Bot) Start() {
	log.Printf("Authorized on account %s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if !b.allowedUsers[update.Message.From.ID] {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ You are not authorized to use this bot.")
			b.api.Send(msg)
			continue
		}

		b.handler.HandleCommand(b.api, update.Message)
	}
}
