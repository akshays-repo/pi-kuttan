Nice! Since you're targeting the **Raspberry Pi Zero 2 W (64-bit)**, I’ll update the PRD accordingly with considerations for its limited resources and ARM64 architecture.

---

## 📄 Product Requirements Document (PRD)  
**Product Name:** GoBotSysMon  
**Version:** 0.1  
**Owner:** [Your Name]  
**Target Platform:** Raspberry Pi Zero 2 W (64-bit)  
**Date:** April 13, 2025

---

### 🧩 Objective  
Build a lightweight and secure **Telegram bot in Go** that runs on a **Raspberry Pi Zero 2 W (64-bit)**. It will respond to two commands—`/help` and `/status`—with restricted access to authorized users only.

---

### 🚀 Features

#### ✅ 1. Command: `/help`  
- **Function:** Displays available commands.  
- **Output Example:**
  ```
  Available Commands:
  /help - Show this message
  /status - Show current CPU and RAM usage
  ```

---

#### ✅ 2. Command: `/status`  
- **Function:** Fetches and displays:
  - Current **CPU usage** (%)
  - Current **RAM usage** (used/total)
- **Output Example:**
  ```
  CPU Usage: 12.3%
  RAM Usage: 210 MB / 512 MB
  ```

---

#### ✅ 3. Access Restriction  
- **Access Control:** Only respond to whitelisted user IDs.
- **Rejection Message for Others:**
  ```
  ❌ You are not authorized to use this bot.
  ```

---

### 🔐 Security Considerations
- **Environment variables** used for:
  - Bot Token
  - Allowed User IDs
- **Minimal dependencies** to reduce attack surface
- No external APIs or databases

---

### 🧠 System Constraints (Pi Zero 2 W)
| Resource | Consideration |
|---------|---------------|
| **CPU** | Quad-core ARM Cortex-A53 – low power |
| **RAM** | 512 MB – avoid memory-intensive operations |
| **Storage** | MicroSD – lightweight binary footprint preferred |
| **OS** | Raspberry Pi OS Lite (64-bit) recommended |

---

### ⚙️ Tech Stack
| Component | Tech |
|----------|------|
| **Language** | Go (compiled for ARM64) |
| **Telegram SDK** | `github.com/go-telegram-bot-api/telegram-bot-api/v5` |
| **System Stats** | `github.com/shirou/gopsutil/v3` (ARM compatible) |
| **Cross-compilation (if needed)** | `GOARCH=arm64 GOOS=linux` |

---

### 📦 Configuration
Store in `.env` or `config.json`:
```env
BOT_TOKEN=your_bot_token
ALLOWED_USER_IDS=123456789,987654321
```

---

### ✅ Acceptance Criteria
| Feature | Criteria |
|--------|---------|
| `/help` | Outputs command list for authorized users |
| `/status` | Returns CPU and RAM info to authorized users |
| Unauthorized Access | Proper denial message |
| Resource Usage | CPU usage <5% idle, memory <30MB usage |

---

### 📅 Future Scope
- Add command `/disk` and `/uptime`
- System alerts via Telegram
- Scheduled status reports
- Battery/power monitoring (if on battery)

---

Would you like a `main.go` scaffold tailored for the Pi Zero 2 W? I can include proper ARM64 cross-compile settings too if you’re not building directly on the Pi.