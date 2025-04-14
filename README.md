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
  - Network details monitoring
- Enhanced Reminder System:
  - Persistent reminders (survives bot restarts)
  - Customizable intervals
  - Multiple reminder types
  - Pause/Resume functionality
  - Status tracking
  - Emoji-enhanced notifications 🎯

### Available Commands

#### 📊 System Monitoring
- `/help` - Display available commands
- `/status` - Show current CPU and RAM usage
- `/temp` - Show CPU temperature
- `/uptime` - Show system uptime
- `/top` - Show top 5 processes
- `/disk` - Show disk usage
- `/network_details` - Show network information

#### ⏰ Reminder Management
Create and Manage Reminders:
- `/reminder_create <type> <interval> <message>` - Create a new reminder
- `/reminder_list` - Show all your active reminders
- `/reminder_pause <id>` - Pause a reminder
- `/reminder_resume <id>` - Resume a paused reminder
- `/reminder_delete <id>` - Delete a reminder
- `/reminder_update <id> <new_interval>` - Update reminder interval

Quick Reminders:
- `/reminder_eye_drop` - Start eye drop reminders (every 2 hours)
- `/reminder_eye_drop_stop` - Stop eye drop reminders
- `/reminder_water` - Start water reminders (every 2 hours)
- `/reminder_water_stop` - Stop water reminders

#### 📝 Example Usage

1. **System Monitoring**
```bash
# Check system status
/status
→ CPU Usage: 23.5%
  RAM Usage: 412 MB / 1024 MB

# Check CPU temperature
/temp
→ CPU Temperature: 45.2°C

# View top processes
/top
→ Top 5 Processes:
  1. mypibot-go (PID: 1234)
     CPU: 2.1%, MEM: 1.5%
  2. systemd (PID: 1)
     CPU: 0.5%, MEM: 0.8%
```

2. **Creating Reminders**
```bash
# Create a water reminder every 90 minutes
/reminder_create water 90 "Time to drink water! 💧"
→ ✅ Reminder created! ID: 1
   Type: water
   Interval: 90 minutes
   Next trigger: 14:30

# Create a medicine reminder every 6 hours
/reminder_create meds 360 "Take your medicine! 💊"
→ ✅ Reminder created! ID: 2
   Type: meds
   Interval: 360 minutes
   Next trigger: 18:00

# Create a custom reminder
/reminder_create custom 45 "Stand up and stretch! 🧘‍♂️"
→ ✅ Reminder created! ID: 3
   Type: custom
   Interval: 45 minutes
   Next trigger: 13:15
```

3. **Managing Reminders**
```bash
# List all active reminders
/reminder_list
→ Active Reminders:
   1. Water (every 90 min)
      Next: 14:30
   2. Medicine (every 360 min)
      Next: 18:00
   3. Stretch (every 45 min)
      Next: 13:15

# Pause a reminder
/reminder_pause 2
→ ⏸️ Reminder #2 (Medicine) paused

# Resume a reminder
/reminder_resume 2
→ ▶️ Reminder #2 (Medicine) resumed
   Next trigger: 18:00

# Update reminder interval
/reminder_update 1 120
→ ⚡ Reminder #1 (Water)
   Interval updated: 90 → 120 minutes
   Next trigger: 15:00

# Delete a reminder
/reminder_delete 3
→ 🗑️ Reminder #3 (Stretch) deleted
```

4. **Quick Reminders**
```bash
# Start eye drop reminders (2-hour intervals)
/reminder_eye_drop
→ 👁️ Eye drop reminders started!
   Next reminder in 2 hours

# Start water reminders (2-hour intervals)
/reminder_water
→ 💧 Water reminders started!
   Next reminder in 2 hours

# Stop specific reminders
/reminder_eye_drop_stop
→ ✅ Eye drop reminders stopped

/reminder_water_stop
→ ✅ Water reminders stopped
```

5. **Network Information**
```bash
# Check network details
/network_details
→ Network Information:
   Interface: wlan0
   IP: 192.168.1.100
   Speed: 100Mbps
   Packets Sent: 1234
   Packets Received: 5678
```

6. **Disk Usage**
```bash
# Check disk space
/disk
→ Disk Usage:
   /: 15.2GB/32GB (47.5%)
   /home: 5.1GB/10GB (51%)
```

#### 💡 Tips
- Use quotes for messages with spaces
- Intervals are in minutes (e.g., 120 = 2 hours)
- Reminders persist after bot restarts
- Use /help to see all available commands
- Quick reminders are preset to 2-hour intervals
- Custom reminders can have any interval

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