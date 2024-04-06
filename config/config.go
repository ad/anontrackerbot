package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

const ConfigFileName = "/data/options.json"

// Config ...
type Config struct {
	TelegramToken        string  `json:"TELEGRAM_TOKEN"`
	TelegramAdminIDs     string  `json:"TELEGRAM_ADMIN_IDS"`
	TelegramAdminIDsList []int64 `json:"-"`

	UPDATE_MESSAGES   string    `json:"UPDATE_MESSAGES"`
	UpdateMessageList []Message `json:"-"`

	DATA_URL string `json:"DATA_URL"`

	UPDATE_DELAY int `json:"UPDATE_DELAY"`

	Debug bool `json:"DEBUG"`
}

type Message struct {
	ChatID          int64
	MessageThreadID int64
	MessageID       int64
}

func InitConfig(args []string) (*Config, error) {
	var config = &Config{
		TelegramToken:        "",
		TelegramAdminIDs:     "",
		TelegramAdminIDsList: []int64{},

		Debug: false,
	}

	var initFromFile = false

	if _, err := os.Stat(ConfigFileName); err == nil {
		jsonFile, err := os.Open(ConfigFileName)
		if err == nil {
			byteValue, _ := io.ReadAll(jsonFile)
			if err = json.Unmarshal(byteValue, &config); err == nil {
				initFromFile = true
			} else {
				fmt.Printf("error on unmarshal config from file %s\n", err.Error())
			}
		}
	}

	if !initFromFile {
		flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
		flags.StringVar(&config.TelegramToken, "telegramToken", lookupEnvOrString("TELEGRAM_TOKEN", config.TelegramToken), "TELEGRAM_TOKEN")
		flags.StringVar(&config.TelegramAdminIDs, "telegramAdminIDs", lookupEnvOrString("TELEGRAM_ADMIN_IDS", config.TelegramAdminIDs), "TELEGRAM_ADMIN_IDS")

		flags.StringVar(&config.UPDATE_MESSAGES, "updateMessages", lookupEnvOrString("UPDATE_MESSAGES", config.UPDATE_MESSAGES), "UPDATE_MESSAGES")

		flags.StringVar(&config.DATA_URL, "dataUrl", lookupEnvOrString("DATA_URL", config.DATA_URL), "DATA_URL")

		flags.IntVar(&config.UPDATE_DELAY, "updateDelay", lookupEnvOrInt("UPDATE_DELAY", config.UPDATE_DELAY), "UPDATE_DELAY")

		flags.BoolVar(&config.Debug, "debug", lookupEnvOrBool("DEBUG", config.Debug), "Debug")

		if err := flags.Parse(args[1:]); err != nil {
			return nil, err
		}
	}

	if config.TelegramAdminIDs != "" {
		chatIDS := strings.Split(config.TelegramAdminIDs, ",")
		for _, chatID := range chatIDS {
			if chatIDInt, err := strconv.ParseInt(strings.Trim(chatID, "\n\t "), 10, 64); err == nil {
				config.TelegramAdminIDsList = append(config.TelegramAdminIDsList, chatIDInt)
			}
		}
	}

	if config.UPDATE_MESSAGES != "" {
		messages := strings.Split(config.UPDATE_MESSAGES, ",")
		for _, message := range messages {
			messageParts := strings.Split(message, ":")
			if len(messageParts) == 3 {
				chatID, _ := strconv.ParseInt(messageParts[0], 10, 64)
				messageThreadID, _ := strconv.ParseInt(messageParts[1], 10, 64)
				messageID, _ := strconv.ParseInt(messageParts[2], 10, 64)
				config.UpdateMessageList = append(config.UpdateMessageList, Message{
					ChatID:          chatID,
					MessageThreadID: messageThreadID,
					MessageID:       messageID,
				})
			}
		}
	}

	return config, nil
}
