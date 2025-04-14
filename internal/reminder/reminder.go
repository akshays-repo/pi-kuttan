package reminder

import (
	"fmt"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Manager struct {
	sync.Mutex
	timers   map[int64]*time.Timer
	eyedrops map[int64]bool
	water    map[int64]bool
	bot      *tgbotapi.BotAPI
}

func NewManager(bot *tgbotapi.BotAPI) *Manager {
	return &Manager{
		timers:   make(map[int64]*time.Timer),
		eyedrops: make(map[int64]bool),
		water:    make(map[int64]bool),
		bot:      bot,
	}
}

const (
	interval = 2 * time.Hour
)

func (m *Manager) StartEyeDrops(chatID int64) string {
	m.Lock()
	defer m.Unlock()

	m.eyedrops[chatID] = true
	
	if _, exists := m.timers[chatID]; !exists {
		timer := time.NewTimer(interval)
		m.timers[chatID] = timer
		go m.reminderLoop(chatID, timer)
	}

	return fmt.Sprintf("✅ Eye drop reminders started! You'll receive notifications every %s.", interval)
}

func (m *Manager) StopEyeDrops(chatID int64) string {
	m.Lock()
	defer m.Unlock()

	if _, exists := m.eyedrops[chatID]; exists {
		delete(m.eyedrops, chatID)
		
		// Only stop timer if both reminders are inactive
		if !m.water[chatID] {
			if timer, exists := m.timers[chatID]; exists {
				timer.Stop()
				delete(m.timers, chatID)
			}
		}
		return "✅ Eye drop reminders stopped."
	}

	return "No active eye drop reminders to stop."
}

func (m *Manager) StartWater(chatID int64) string {
	m.Lock() 
	defer m.Unlock()

	m.water[chatID] = true

	if _, exists := m.timers[chatID]; !exists {
		timer := time.NewTimer(0)
		m.timers[chatID] = timer
		go m.reminderLoop(chatID, timer)
	}

	return fmt.Sprintf("✅ Water reminders started! You'll receive notifications every %s.", interval)
}

func (m *Manager) StopWater(chatID int64) string {
	m.Lock()
	defer m.Unlock()

	if _, exists := m.water[chatID]; exists {
		delete(m.water, chatID)
		
		// Only stop timer if both reminders are inactive
		if !m.eyedrops[chatID] {
			if timer, exists := m.timers[chatID]; exists {
				timer.Stop()
				delete(m.timers, chatID)
			}
		}
		return "✅ Water reminders stopped."
	}

	return "No active water reminders to stop."
}

func (m *Manager) Status(chatID int64) string {
	m.Lock()
	defer m.Unlock()

	if _, exists := m.timers[chatID]; !exists {
		return "No active reminders."
	}

	status := "Current Reminder Status:\n"
	if m.eyedrops[chatID] {
		status += "👁️ Eye Drops: Active\n"
	} else {
		status += "👁️ Eye Drops: Inactive\n"
	}
	if m.water[chatID] {
		status += "💧 Water: Active\n"
	} else {
		status += "💧 Water: Inactive\n"
	}
	return status
}

func (m *Manager) reminderLoop(chatID int64, timer *time.Timer) {
	for {
		select {
		case <-timer.C:
			m.Lock()
			if m.eyedrops[chatID] {
				msg := tgbotapi.NewMessage(chatID, "👁️ Time for eye drops! Take care of your eyes.")
				m.bot.Send(msg)
			}
			if m.water[chatID] {
				msg := tgbotapi.NewMessage(chatID, "💧 Drink water! Stay hydrated.")
				m.bot.Send(msg)
			}
			timer.Reset(interval)
			m.Unlock()
		}
	}
}
