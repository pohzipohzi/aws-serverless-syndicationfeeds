package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/go-multierror"
	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//nolint:gochecknoglobals
var (
	envUrls             = os.Getenv("URLS")
	envDdbTableName     = os.Getenv("DDB_TABLE_NAME")
	envTelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN") //nolint:gosec
	envTelegramChatID   = os.Getenv("TELEGRAM_CHAT_ID")
)

type ItemHandler interface {
	fmt.Stringer
	Handle(*gofeed.Item) error
}

func Main() error {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	handlers := getItemHandlers()
	if len(handlers) == 0 {
		log.Fatal().Msg("no handler configured")
	}

	if envDdbTableName == "" {
		log.Fatal().Msg("ddb table name not configured")
	}

	var errors *multierror.Error
	urls := []string{}
	if err := json.Unmarshal([]byte(envUrls), &urls); err != nil {
		return err
	}

	fp := gofeed.NewParser()
	sess, err := session.NewSession()
	if err != nil {
		errors = multierror.Append(errors, err)
	}
	ddb := dynamodb.New(sess)
	for _, u := range urls {
		feed, err := fp.ParseURL(u)
		if err != nil {
			errors = multierror.Append(errors, err)
			continue
		}
		log.Info().Int("len_items", len(feed.Items)).Str("url", u).Msg("successfully parsed feed url")
		sort.Sort(feed) // sort items by publish time
		ddbItems, err := ddbGetItems(ddb, feed)
		if err != nil {
			errors = multierror.Append(errors, err)
			continue
		}
		for _, i := range feed.Items {
			if !shouldHandleItem(feed, i, ddbItems) {
				log.Info().Str("guid", i.GUID).Msg("skipped handling item")
				continue
			}
			err = ddbUpdateItem(ddb, feed, i)
			if err != nil {
				errors = multierror.Append(errors, err)
			}
			for _, h := range handlers {
				err = h.Handle(i)
				if err != nil {
					errors = multierror.Append(errors, err)
				} else {
					log.Info().Stringer("handler", h).Msg("successfully handled item")
				}
			}
		}
	}

	return errors.ErrorOrNil()
}

func getItemHandlers() []ItemHandler {
	ret := []ItemHandler{}
	if envTelegramBotToken != "" && envTelegramChatID != "" {
		ret = append(ret, &Telegram{
			token:  envTelegramBotToken,
			chatID: envTelegramChatID,
		})
	}
	return ret
}
