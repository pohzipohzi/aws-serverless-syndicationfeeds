package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"lambda-feed-notifier/internal/handlers"
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
	envUrls         = os.Getenv("URLS")
	envDdbTableName = os.Getenv("DDB_TABLE_NAME")
)

func Main() error {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	ha := handlers.All()
	if len(ha) == 0 {
		return errors.New("no handler configured")
	}

	if envDdbTableName == "" {
		return errors.New("ddb table name not configured")
	}

	urls := []string{}
	if err := json.Unmarshal([]byte(envUrls), &urls); err != nil {
		return fmt.Errorf("failed to unmarshal urls: %w", err)
	}

	var errs *multierror.Error
	fp := gofeed.NewParser()
	sess, err := session.NewSession()
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	ddb := dynamodb.New(sess)
	for _, u := range urls {
		feed, err := fp.ParseURL(u)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		log.Info().Int("len_items", len(feed.Items)).Str("url", u).Msg("successfully parsed feed url")
		sort.Sort(feed) // sort items by publish time
		ddbItems, err := ddbGetItems(ddb, feed)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
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
	}

	return errs.ErrorOrNil()
}
