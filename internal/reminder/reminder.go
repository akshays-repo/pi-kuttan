package reminder

import (
	"fmt"
	"log"
	"mypibot-go/internal/storage"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Manager struct {
	sync.Mutex
	db        *storage.Database
	timers    map[int64]*time.Timer
	bot       *tgbotapi.BotAPI
}

func NewManager(db *storage.Database, bot *tgbotapi.BotAPI) *Manager {
	return &Manager{
		db:     db,
		timers: make(map[int64]*time.Timer),
		bot:    bot,
	}
}


// CreateReminder creates a new reminder and starts its timer
func (m *Manager) CreateReminder(chatID int64, interval int, message string) (int64, error) {
	m.Lock()
	defer m.Unlock()


	// Create reminder in database
	reminder, err := m.db.CreateReminder(chatID, interval, message)
	if err != nil {
		return 0, fmt.Errorf("failed to create reminder: %w", err)
	}

	// Start timer
	timer := time.NewTimer(time.Duration(interval) * time.Minute)
	m.timers[reminder.ID] = timer

	// Start reminder loop in background
	go m.reminderLoop(reminder.ID, timer, time.Duration(interval)*time.Minute, message)

	return reminder.ID, nil
}

// ListReminders returns all active reminders for a chat
func (m *Manager) ListReminders(chatID int64) ([]*storage.Reminder, error) {
	reminders, err := m.db.ListActiveReminders(chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to list reminders: %w", err)
	}
	return reminders, nil
}

// PauseReminder pauses a reminder
func (m *Manager) PauseReminder(chatID int64, reminderID int64) error {
	m.Lock()
	defer m.Unlock()

	// Verify reminder belongs to this chat
	reminder, err := m.db.GetReminder(reminderID)
	if err != nil {
		return fmt.Errorf("failed to get reminder: %w", err)
	}
	if reminder.ChatID != chatID {
		return fmt.Errorf("reminder not found")
	}

	// Stop the timer
	if timer, exists := m.timers[reminderID]; exists {
		timer.Stop()
		delete(m.timers, reminderID)
	}

	// Update status in database
	if err := m.db.UpdateReminderStatus(reminderID, "paused"); err != nil {
		return fmt.Errorf("failed to pause reminder: %w", err)
	}

	return nil
}

// ResumeReminder resumes a paused reminder
func (m *Manager) ResumeReminder(chatID int64, reminderID int64) error {
	m.Lock()
	defer m.Unlock()

	// Verify reminder belongs to this chat
	reminder, err := m.db.GetReminder(reminderID)
	if err != nil {
		return fmt.Errorf("failed to get reminder: %w", err)
	}
	if reminder.ChatID != chatID {
		return fmt.Errorf("reminder not found")
	}

	// Create new timer
	timer := time.NewTimer(time.Duration(reminder.Interval) * time.Minute)
	m.timers[reminderID] = timer

	// Start reminder loop
	go m.reminderLoop(reminderID, timer, time.Duration(reminder.Interval)*time.Minute, reminder.Message)

	// Update status in database
	if err := m.db.UpdateReminderStatus(reminderID, "active"); err != nil {
		return fmt.Errorf("failed to resume reminder: %w", err)
	}

	return nil
}

// DeleteReminder deletes a reminder
func (m *Manager) DeleteReminder(chatID int64, reminderID int64) error {
	m.Lock()
	defer m.Unlock()

	// Verify reminder belongs to this chat
	reminder, err := m.db.GetReminder(reminderID)
	if err != nil {
		return fmt.Errorf("failed to get reminder: %w", err)
	}
	if reminder.ChatID != chatID {
		return fmt.Errorf("reminder not found")
	}

	// Stop and remove timer
	if timer, exists := m.timers[reminderID]; exists {
		timer.Stop()
		delete(m.timers, reminderID)
	}

	// Delete from database
	if err := m.db.DeleteReminder(reminderID); err != nil {
		return fmt.Errorf("failed to delete reminder: %w", err)
	}

	return nil
}

// UpdateInterval updates a reminder's interval
func (m *Manager) UpdateInterval(reminderID int64, chatID int64, newInterval int) error {
	m.Lock()
	defer m.Unlock()

	// Verify reminder belongs to this chat
	reminder, err := m.db.GetReminder(reminderID)
	if err != nil {
		return fmt.Errorf("failed to get reminder: %w", err)
	}
	if reminder.ChatID != chatID {
		return fmt.Errorf("reminder not found")
	}

	// Update in database
	if err := m.db.UpdateReminderInterval(reminderID, newInterval); err != nil {
		return fmt.Errorf("failed to update interval: %w", err)
	}

	// Restart timer with new interval
	if timer, exists := m.timers[reminderID]; exists {
		timer.Stop()
		newTimer := time.NewTimer(time.Duration(newInterval) * time.Minute)
		m.timers[reminderID] = newTimer
		go m.reminderLoop(reminderID, newTimer, time.Duration(newInterval)*time.Minute, reminder.Message)
	}

	return nil
}

// RecoverActiveReminders recovers all active reminders on startup
func (m *Manager) RecoverActiveReminders() error {
	m.Lock()
	defer m.Unlock()

	log.Println("Starting reminder recovery...")

	// Get all active reminders from database
	reminders, err := m.db.GetAllActiveReminders()
	if err != nil {
		return fmt.Errorf("failed to recover reminders: %w", err)
	}

	recoveredCount := 0
	for _, reminder := range reminders {
		// Calculate next trigger time
		var nextTrigger time.Time
		if reminder.NextTrigger.Valid {
			nextTrigger = reminder.NextTrigger.Time
		} else {
		// If next_trigger is not set, calculate from last_triggered or created_at
			if reminder.LastTriggered.Valid {
				nextTrigger = reminder.LastTriggered.Time.Add(time.Duration(reminder.Interval) * time.Minute)
			} else {
				nextTrigger = reminder.CreatedAt.Add(time.Duration(reminder.Interval) * time.Minute)
			}
		}

		// Calculate duration until next trigger
		duration := time.Until(nextTrigger)
		if duration < 0 {
			// If we're past the trigger time, schedule for next interval
			duration = time.Duration(reminder.Interval) * time.Minute
		}

		// Create new timer
		timer := time.NewTimer(duration)
		m.timers[reminder.ID] = timer

		// Start reminder loop
		go m.reminderLoop(reminder.ID, timer, time.Duration(reminder.Interval)*time.Minute, reminder.Message)
		
		recoveredCount++
		log.Printf("Recovered reminder ID %d, type: %s, next trigger in: %.2f minutes", 
			reminder.ID, reminder.Type, duration.Minutes())
	}

	log.Printf("Reminder recovery completed. Recovered %d active reminders", recoveredCount)
	return nil
}

// reminderLoop handles the periodic reminder notifications
func (m *Manager) reminderLoop(reminderID int64, timer *time.Timer, interval time.Duration, message string) {
	for {
		select {
		case <-timer.C:
			m.Lock()
			// Get current reminder status
			reminder, err := m.db.GetReminder(reminderID)
			if err != nil {
				log.Printf("Error getting reminder %d: %v", reminderID, err)
				m.Unlock()
				return
			}

			if reminder.Status != "active" {
				log.Printf("Reminder %d is no longer active, stopping loop", reminderID)
				delete(m.timers, reminderID)
				m.Unlock()
				return
			}

			// Send notification
			msg := tgbotapi.NewMessage(reminder.ChatID, fmt.Sprintf("🔔 Reminder: %s", message))
			_, err = m.bot.Send(msg)
			if err != nil {
				log.Printf("Error sending reminder %d: %v", reminderID, err)
			}

			// Update last triggered time in database
			err = m.db.UpdateReminderTrigger(reminderID)
			if err != nil {
				log.Printf("Error updating reminder trigger %d: %v", reminderID, err)
			}

			// Reset timer for next interval
			timer.Reset(interval)
			m.Unlock()
		}
	}
}