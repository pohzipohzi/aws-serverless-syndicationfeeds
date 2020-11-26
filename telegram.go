package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/mmcdole/gofeed"
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
	get := fmt.Sprintf(telegramSendMessageEndpoint, h.token, h.chatID, url.QueryEscape(text))
	rc, err := http.Get(get)
	if err != nil {
		return err
	}
	defer rc.Body.Close()
	b, err := ioutil.ReadAll(rc.Body)
	if err != nil {
		return err
	}
	log.Println(string(b))
	return nil
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}
