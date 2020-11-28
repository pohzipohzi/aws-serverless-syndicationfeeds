package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_All(t *testing.T) {
	envTelegramBotToken = "telegrambottoken"
	envTelegramChatID = "telegramchatid"
	assert.Equal(t, 1, len(All()))
}
