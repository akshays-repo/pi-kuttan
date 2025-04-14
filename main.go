package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/shirou/gopsutil/v4/sensors"
)

var (
	botToken     string
	allowedUsers []int64
)

func init() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	botToken = os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN is required")
	}

	// Parse allowed user IDs
	allowedUserIDs := os.Getenv("ALLOWED_USER_IDS")
	if allowedUserIDs == "" {
		log.Fatal("ALLOWED_USER_IDS is required")
	}

	for _, id := range strings.Split(allowedUserIDs, ",") {
		userID, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
		if err != nil {
			log.Printf("Warning: Invalid user ID %s: %v", id, err)
			continue
		}
		allowedUsers = append(allowedUsers, userID)
	}
}

func isUserAllowed(userID int64) bool {
	for _, id := range allowedUsers {
		if id == userID {
			return true
		}
	}
	return false
}

func getSystemStats() (string, error) {
	// Get CPU usage
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return "", fmt.Errorf("error getting CPU usage: %v", err)
	}

	// Get memory usage
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return "", fmt.Errorf("error getting memory usage: %v", err)
	}

	return fmt.Sprintf("CPU Usage: %.1f%%\nRAM Usage: %d MB / %d MB",
		cpuPercent[0],
		memInfo.Used/1024/1024,
		memInfo.Total/1024/1024), nil
}
func getTemperature() (string, error) {
	temps, err := sensors.SensorsTemperatures()
	if err != nil {
		return "", fmt.Errorf("error getting temperature: %v", err)
	}

	var cpuTemp float64
	for _, temp := range temps {
		if strings.Contains(strings.ToLower(temp.SensorKey), "cpu") {
			cpuTemp = temp.Temperature
			break
		}
	}

	return fmt.Sprintf("CPU Temperature: %.1f°C", cpuTemp), nil
}

func getUptime() (string, error) {
	uptime, err := host.Uptime()
	if err != nil {
		return "", fmt.Errorf("error getting uptime: %v", err)
	}

	duration := time.Duration(uptime) * time.Second
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	return fmt.Sprintf("System Uptime: %d days, %d hours, %d minutes", days, hours, minutes), nil
}

func getTopProcesses() (string, error) {
	processes, err := process.Processes()
	if err != nil {
		return "", fmt.Errorf("error getting processes: %v", err)
	}

	type ProcessInfo struct {
		pid     int32
		name    string
		cpuPerc float64
		memPerc float32
	}

	var processInfos []ProcessInfo
	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}

		cpuPerc, err := p.CPUPercent()
		if err != nil {
			continue
		}

		memPerc, err := p.MemoryPercent()
		if err != nil {
			continue
		}

		processInfos = append(processInfos, ProcessInfo{
			pid:     p.Pid,
			name:    name,
			cpuPerc: cpuPerc,
			memPerc: memPerc,
		})
	}

	// Sort by CPU usage
	sort.Slice(processInfos, func(i, j int) bool {
		return processInfos[i].cpuPerc > processInfos[j].cpuPerc
	})

	// Get top 5 processes
	var result strings.Builder
	result.WriteString("Top 5 Processes:\n")
	for i := 0; i < len(processInfos) && i < 5; i++ {
		result.WriteString(fmt.Sprintf("%d. %s (PID: %d)\n   CPU: %.1f%%, MEM: %.1f%%\n",
			i+1,
			processInfos[i].name,
			processInfos[i].pid,
			processInfos[i].cpuPerc,
			processInfos[i].memPerc,
		))
	}

	return result.String(), nil
}

func getDiskUsage() (string, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return "", fmt.Errorf("error getting disk partitions: %v", err)
	}

	var result strings.Builder
	result.WriteString("Disk Usage:\n")

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}
		result.WriteString(fmt.Sprintf("%s:\n  Used: %.1f GB / %.1f GB (%.1f%%)\n",
			partition.Mountpoint,
			float64(usage.Used)/1024/1024/1024,
			float64(usage.Total)/1024/1024/1024,
			usage.UsedPercent,
		))
	}

	return result.String(), nil
}

func main() {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Check if user is authorized
		if !isUserAllowed(update.Message.From.ID) {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ You are not authorized to use this bot.")
			bot.Send(msg)
			continue
		}

		// Handle commands
		switch update.Message.Command() {
		case "help":
			helpText := `Available Commands:
/help - Show this message
/status - Show current CPU and RAM usage
/temp - Show CPU temperature
/uptime - Show system uptime
/top - Show top 5 processes
/disk - Show disk usage`
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
			bot.Send(msg)

		case "status":
			stats, err := getSystemStats()
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error getting system stats: "+err.Error())
				bot.Send(msg)
				continue
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, stats)
			bot.Send(msg)

		case "temp":
			temp, err := getTemperature()
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error getting temperature: "+err.Error())
				bot.Send(msg)
				continue
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, temp)
			bot.Send(msg)

		case "uptime":
			uptime, err := getUptime()
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error getting uptime: "+err.Error())
				bot.Send(msg)
				continue
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, uptime)
			bot.Send(msg)

		case "top":
			processes, err := getTopProcesses()
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error getting processes: "+err.Error())
				bot.Send(msg)
				continue
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, processes)
			bot.Send(msg)

		case "disk":
			disk, err := getDiskUsage()
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error getting disk usage: "+err.Error())
				bot.Send(msg)
				continue
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, disk)
			bot.Send(msg)
		}
	}
} 