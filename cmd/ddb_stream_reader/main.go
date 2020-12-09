package main

import (
	"aws-serverless-syndicationfeeds/cmd/adapter"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(handler)
}

func handler(e events.DynamoDBEvent) error {
	for _, r := range e.Records {
		item := adapter.DdbImageToFeedItem(r.Change.NewImage)
		log.Printf("Received feed item: %v\n", item)
	}
	return nil
}
