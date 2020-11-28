package handlers

import (
	"fmt"
	"lambda-feed-notifier/internal/handlers/telegram"
	"os"

	"github.com/mmcdole/gofeed"
)

//nolint:gochecknoglobals
var (
	envTelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN") //nolint:gosec
	envTelegramChatID   = os.Getenv("TELEGRAM_CHAT_ID")
)

type Handler interface {
	fmt.Stringer
	Handle(*gofeed.Item) error
}

func All() []Handler {
	ret := []Handler{}
	if envTelegramBotToken != "" && envTelegramChatID != "" {
		ret = append(ret, telegram.New(envTelegramBotToken, envTelegramChatID))
	}
	return ret
}
