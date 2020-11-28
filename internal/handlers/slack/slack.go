package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	url string
}

type slackMessage struct {
	Text string `json:"text"`
}

func New(url string) *Handler {
	return &Handler{
		url: url,
	}
}

func (h *Handler) String() string {
	return "slack"
}

func (h *Handler) Handle(i *gofeed.Item) error {
	var text string
	text += fmt.Sprintf("Title: %s\n", i.Title)
	text += fmt.Sprintf("Description: %s\n", truncate(i.Description, 100))
	text += fmt.Sprintf("Published: %s\n", i.Published)
	text += fmt.Sprintf("Link: %s\n", i.Link)
	m := slackMessage{Text: text}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, h.url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Add("Content-type", "application/json")
	rc, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer rc.Body.Close()
	b, err = ioutil.ReadAll(rc.Body)
	if err != nil {
		return err
	}
	log.Info().Bytes("response_body", b).Int("response_code", rc.StatusCode).Msg("successfully sent message to slack")
	return nil
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}
