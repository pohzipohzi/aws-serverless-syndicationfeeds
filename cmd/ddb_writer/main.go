package main

import (
	"aws-serverless-syndicationfeeds/cmd/adapter"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mmcdole/gofeed"
)

const envDdbTableName = "DDB_TABLE_NAME"

type input struct {
	URL string `json:"url"`
}

func main() {
	lambda.Start(handler)
}

func handler(in input) error {
	ddbTableName := os.Getenv(envDdbTableName)
	if ddbTableName == "" {
		return errors.New("ddb table name not configured")
	}
	fp := gofeed.NewParser()
	sess, err := session.NewSession()
	if err != nil {
		return fmt.Errorf("failed to start new aws session: %w", err)
	}
	ddb := dynamodb.New(sess)
	feed, err := fp.ParseURL(in.URL)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}
	log.Printf("retrieved %d items\n", len(feed.Items))
	for _, i := range feed.Items {
		ddbUpdateItemInput := adapter.FeedItemToDdbUpdateItemInput(feed, i, ddbTableName)
		_, err = ddb.UpdateItem(ddbUpdateItemInput)
		if err != nil {
			log.Printf("failed to write %s: %v\n", i.GUID, err)
		} else {
			log.Printf("successfully wrote %s", i.GUID)
		}
	}
	return nil
}
