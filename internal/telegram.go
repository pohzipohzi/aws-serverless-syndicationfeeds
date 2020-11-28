package internal

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"
)

const (
	telegramSendMessageEndpoint = "https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s"
)

type Telegram struct {
	token  string
	chatID string
}

func (h *Telegram) String() string {
	return "telegram"
}

func (h *Telegram) Handle(i *gofeed.Item) error {
	var text string
	text += fmt.Sprintf("Title: %s\n", i.Title)
	text += fmt.Sprintf("Description: %s\n", truncate(i.Description, 100))
	text += fmt.Sprintf("Published: %s\n", i.Published)
	text += fmt.Sprintf("Link: %s\n", i.Link)
	getURL := fmt.Sprintf(telegramSendMessageEndpoint, h.token, h.chatID, url.QueryEscape(text))
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, getURL, nil)
	if err != nil {
		return err
	}
	rc, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer rc.Body.Close()
	b, err := ioutil.ReadAll(rc.Body)
	if err != nil {
		return err
	}
	log.Info().Bytes("response", b).Msg("successfully sent message to telegram")
	return nil
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}
