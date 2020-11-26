package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/go-multierror"
	"github.com/mmcdole/gofeed"
)

//nolint:gochecknoglobals
var (
	envUrls             = os.Getenv("URLS")
	envTelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN") //nolint:gosec
	envTelegramChatID   = os.Getenv("TELEGRAM_CHAT_ID")
)

type ItemHandler interface {
	fmt.Stringer
	Handle(*gofeed.Item) error
}

func main() {
	lambda.Start(lambdaHandler)
}

func lambdaHandler() error {
	handlers := getItemHandlers()
	if len(handlers) == 0 {
		log.Println("no handler configured")
		return nil
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
		log.Printf("found %d items for url %s\n", len(feed.Items), u)
		sort.Sort(feed) // sort items by publish time
		ddbItems, err := ddbGetItems(ddb, feed)
		if err != nil {
			errors = multierror.Append(errors, err)
			continue
		}
		for _, i := range feed.Items {
			if !shouldHandleItem(feed, i, ddbItems) {
				log.Printf("skipped handling item %s\n", i.GUID)
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
					log.Printf("%s handler successfully completed\n", h)
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
