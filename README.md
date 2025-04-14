# 🤖 PiKuttan (പൈ-കുട്ടൻ )

A cute little system monitoring bot running on my Raspberry Pi Zero 2 W! 🍓 This tiny but mighty single-board computer helps me keep track of system stats and sends helpful reminders throughout the day.


A lightweight Telegram bot built in Go that provides system monitoring and reminder functionality.

## Features
- System monitoring:
  - CPU and RAM usage monitoring
  - CPU temperature tracking
  - System uptime display
  - Top 5 processes view
  - Disk usage monitoring
- Reminder system:
  - Eye drop reminders every 2 hours
  - Water consumption reminders every 2 hours
  - Concurrent reminder management for multiple users
  - Thread-safe operations
  - Emoji-enhanced notifications 🎯

### Available Commands
- `/help` - Display available commands
- `/status` - Show current CPU and RAM usage
- `/temp` - Show CPU temperature
- `/uptime` - Show system uptime
- `/top` - Show top 5 processes
- `/disk` - Show disk usage
- `/reminder_eye_drop` - Start eye drop reminders
- `/reminder_eye_drop_stop` - Stop eye drop reminders
- `/reminder_water` - Start water reminders
- `/reminder_water_stop` - Stop water reminders

## Requirements

- Go 1.21 or later
- Telegram Bot Token

## Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/akshays-repo/pi-kuttan.git
   cd pi-kuttan
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Create your `.env` file:
   ```bash
   cp .env.example .env
   ```

4. Edit the `.env` file with your:
   - Telegram Bot Token (from [@BotFather](https://t.me/botfather))
   - Allowed user IDs (comma-separated)

## Building

### For local development