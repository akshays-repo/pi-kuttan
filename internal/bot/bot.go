package bot

import (
	"log"

	"mypibot-go/internal/config"
	"mypibot-go/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api          *tgbotapi.BotAPI
	allowedUsers map[int64]bool
	handler      *Handler
	db           *storage.Database
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

	// Initialize database
	db, err := storage.NewDatabase(cfg.DatabasePath)
	if err != nil {
		return nil, err
	}

	// Create bot instance
	bot := &Bot{
		api:          api,
		allowedUsers: allowedUsers,
		db:           db,
	}

	// Create handler with database
	bot.handler = NewHandler(db, api)

	// Recover active reminders
	if err := bot.recoverReminders(); err != nil {
		log.Printf("Warning: Failed to recover reminders: %v", err)
	}

	return bot, nil
}

func (b *Bot) recoverReminders() error {
	log.Println("Starting reminder recovery process...")

	// Get all active reminders
	reminders, err := b.db.GetAllActiveReminders()
	if err != nil {
		return err
	}

	recoveredCount := 0
	for _, reminder := range reminders {
		b.handler.reminder.ResumeReminder(reminder.ChatID , reminder.ID)
	}

	log.Printf("Reminder recovery completed. Recovered %d active reminders", recoveredCount)
	return nil
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

// Add cleanup method
func (b *Bot) Stop() {
	if b.db != nil {
		b.db.Close()
	}
}
