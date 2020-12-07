package main

import (
	"errors"
	"lambda-feed-notifier/cmd/handlers"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/go-multierror"
	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//nolint:gochecknoglobals
var envDdbTableName = os.Getenv("DDB_TABLE_NAME")

type input struct {
	URL string `json:"url"`
}

func main() {
	lambda.Start(handler)
}

func handler(in input) error {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if envDdbTableName == "" {
		return errors.New("ddb table name not configured")
	}

	ha := handlers.All()
	if len(ha) == 0 {
		log.Info().Int("len_handlers", len(ha)).Msg("configured handlers")
	}

	var errs *multierror.Error
	fp := gofeed.NewParser()
	sess, err := session.NewSession()
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	ddb := dynamodb.New(sess)
	feed, err := fp.ParseURL(in.URL)
	if err != nil {
		errs = multierror.Append(errs, err)
		return errs.ErrorOrNil()
	}
	log.Info().Int("len_items", len(feed.Items)).Str("url", in.URL).Msg("successfully parsed feed url")
	ddbItems, err := ddbGetItems(ddb, feed)
	if err != nil {
		errs = multierror.Append(errs, err)
		return errs.ErrorOrNil()
	}
	for _, i := range feed.Items {
		if !shouldHandleItem(feed, i, ddbItems) {
			log.Info().Str("guid", i.GUID).Msg("skipped handling item")
			continue
		}
		err = ddbUpdateItem(ddb, feed, i)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		for _, h := range ha {
			err = h.Handle(i)
			if err != nil {
				errs = multierror.Append(errs, err)
			} else {
				log.Info().Stringer("handler", h).Msg("successfully handled item")
			}
		}
	}

	return errs.ErrorOrNil()
}
