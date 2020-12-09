package main

import (
	"aws-serverless-syndicationfeeds/cmd/adapter"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mmcdole/gofeed"
)

const envDdbTableName = "DDB_TABLE_NAME"

type handler struct {
	ddbTableName string
	feedParser   *gofeed.Parser
	ddb          *dynamodb.DynamoDB
}

type input struct {
	URL string `json:"url"`
}

func main() {
	ddbTableName := os.Getenv(envDdbTableName)
	if ddbTableName == "" {
		log.Fatal("ddb table name not configured")
	}
	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("failed to start aws session: %v", err)
	}
	lambda.Start((&handler{
		ddbTableName: ddbTableName,
		feedParser:   gofeed.NewParser(),
		ddb:          dynamodb.New(sess),
	}).Handle)
}

func (h *handler) Handle(in input) error {
	feed, err := h.feedParser.ParseURL(in.URL)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}
	log.Printf("retrieved %d items\n", len(feed.Items))
	for _, i := range feed.Items {
		ddbUpdateItemInput := adapter.FeedItemToDdbUpdateItemInput(feed, i, h.ddbTableName)
		_, err = h.ddb.UpdateItem(ddbUpdateItemInput)
		if err != nil {
			log.Printf("failed to write %s: %v\n", i.GUID, err)
		}
	}
	return nil
}
