package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken     string
	AllowedUsers []int64
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		return nil, fmt.Errorf("BOT_TOKEN is required")
	}

	// Parse allowed user IDs
	allowedUserIDs := os.Getenv("ALLOWED_USER_IDS")
	if allowedUserIDs == "" {
		return nil, fmt.Errorf("ALLOWED_USER_IDS is required")
	}

	var allowedUsers []int64
	for _, id := range strings.Split(allowedUserIDs, ",") {
		userID, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID %s: %w", id, err)
		}
		allowedUsers = append(allowedUsers, userID)
	}

	return &Config{
		BotToken:     botToken,
		AllowedUsers: allowedUsers,
	}, nil
}
