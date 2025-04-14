package bot

import (
	"mypibot-go/internal/monitor"
	"mypibot-go/internal/reminder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	monitor  *monitor.Monitor
	reminder *reminder.Manager
}

func NewHandler(bot *tgbotapi.BotAPI) *Handler {
	return &Handler{
		monitor:  monitor.New(),
		reminder: reminder.NewManager(bot),
	}
}

func (h *Handler) HandleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	var text string
	var err error

	switch message.Command() {
	case "help":
		text = `Available Commands:
/help - Show this message
/status - Show current CPU and RAM usage
/temp - Show CPU temperature
/uptime - Show system uptime
/top - Show top 5 processes
/disk - Show disk usage

Reminder Commands:
/reminder_eye_drop - Start eye drop reminders
/reminder_eye_drop_stop - Stop eye drop reminders
/reminder_water - Start water reminders
/reminder_water_stop - Stop water reminders`

	case "status":
		text, err = h.monitor.GetSystemStats()

	case "temp":
		text, err = h.monitor.GetTemperature()

	case "uptime":
		text, err = h.monitor.GetUptime()

	case "top":
		text, err = h.monitor.GetTopProcesses()

	case "disk":
		text, err = h.monitor.GetDiskUsage()

	case "reminder_eye_drop":
		text = h.reminder.StartEyeDrops(message.Chat.ID)

	case "reminder_eye_drop_stop":
		text = h.reminder.StopEyeDrops(message.Chat.ID)

	case "reminder_water":
		text = h.reminder.StartWater(message.Chat.ID)

	case "reminder_water_stop":
		text = h.reminder.StopWater(message.Chat.ID)

	default:
		text = "Unknown command. Use /help to see available commands."
	}

	if err != nil {
		text = "Error: " + err.Error()
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	bot.Send(msg)
}
