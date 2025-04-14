package bot

import (
	"fmt"
	"mypibot-go/internal/monitor"
	"mypibot-go/internal/reminder"
	"mypibot-go/internal/storage"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	monitor  *monitor.Monitor
	reminder *reminder.Manager
}

func NewHandler(db *storage.Database, bot *tgbotapi.BotAPI) *Handler {
	return &Handler{
		monitor:  monitor.New(),
		reminder: reminder.NewManager(db, bot),
	}
}

func (h *Handler) HandleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	var text string
	var err error

	switch message.Command() {
	case "help":
		text := `<b>Bot Commands Guide</b>

<b>📊 System Monitoring</b>
• /help - Show this help message
• /status - Show CPU and RAM usage
• /temp - Show CPU temperature
• /uptime - Show system uptime
• /top - Show top 5 processes
• /disk - Show disk usage
• /network_details - Show network details
• /reboot - Reboot the system (admin only)


<b>⏰ Reminder Commands</b>

<b>Create New Reminder:</b>
/reminder_create  &lt;interval&gt; &lt;message&gt;

<b>Examples:</b>
• /reminder_create water 120 "Drink water! 💧"
• /reminder_create meds 360 "Take medicine! 💊"

<b>Manage Reminders:</b>
• /reminder_list - Show active reminders
• /reminder_pause &lt;id&gt; - Pause a reminder
• /reminder_resume &lt;id&gt; - Resume a reminder
• /reminder_delete &lt;id&gt; - Delete a reminder

<b>Quick Reminders:</b>
• /reminder_eye_drop - Start eye drops (2h)
• /reminder_water - Start water (2h)

<b>💡 Tips:</b>
• Intervals are in minutes
• Use quotes for messages with spaces
• Use /reminder_list to get reminder IDs`
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return

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

	case "network_details":
		text, err = h.monitor.GetNetworkDetails()
		// TODO: Add admin check
	case "reboot":
		text, err = h.monitor.RebootSystem()
		
	case "reminder_create":
		var reminderID int64
		args := strings.TrimSpace(message.CommandArguments())
		if args == "" {
			err = fmt.Errorf("not enough arguments. Usage: /reminder_create <interval> <message>")
		} else {
			// Split only on first space to keep message intact
			parts := strings.SplitN(args, " ", 2)
			if len(parts) < 2 {
				err = fmt.Errorf("not enough arguments. Usage: /reminder_create <interval> <message>")
			} else {
				interval, parseErr := strconv.Atoi(parts[0])
				if parseErr != nil {
					err = fmt.Errorf("invalid interval: %v", parseErr)
				} else {
					reminderMessage := parts[1]
					reminderID, err = h.reminder.CreateReminder(message.Chat.ID, interval, reminderMessage)
					if err == nil {
						text = fmt.Sprintf("Reminder created successfully! ID: %d", reminderID)
					}
				}
			}
		}

	case "reminder_list":
		reminders, err := h.reminder.ListReminders(message.Chat.ID)
		if err == nil {
			text = "Active Reminders:\n"
			for _, reminder := range reminders {
				text += fmt.Sprintf("ID: %d\nInterval: %d minutes\nMessage: %s\n\n",
					reminder.ID, reminder.Interval, reminder.Message)
			}
		}
	case "reminder_pause":
		var reminderID int64
		reminderID, err = strconv.ParseInt(message.CommandArguments(), 10, 64)
		if err == nil {
			err = h.reminder.PauseReminder(message.Chat.ID, reminderID)
		}
	case "reminder_resume":
		var reminderID int64
		reminderID, err = strconv.ParseInt(message.CommandArguments(), 10, 64)
		if err == nil {
			err = h.reminder.ResumeReminder(message.Chat.ID, reminderID)
		}
	case "reminder_delete":
		var reminderID int64
		reminderID, err = strconv.ParseInt(message.CommandArguments(), 10, 64)
		if err == nil {
			err = h.reminder.DeleteReminder(message.Chat.ID, reminderID)
		}
		
	default:
		text = "Unknown command. Use /help to see available commands."
	}

	if err != nil {
		text = "Error: " + err.Error()
	}


	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	bot.Send(msg)
}
