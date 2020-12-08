package main

import (
	schema "aws-serverless-syndicationfeeds/cmd/ddb_schema"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mmcdole/gofeed"
)

func main() {
	lambda.Start(handler)
}

func handler(e events.DynamoDBEvent) error {
	for _, r := range e.Records {
		item := feedItemFromImage(r.Change.NewImage)
		log.Printf("Received feed item: %v\n", item)
	}
	return nil
}

func feedItemFromImage(image map[string]events.DynamoDBAttributeValue) *gofeed.Item {
	if image == nil {
		return nil
	}
	return &gofeed.Item{
		GUID:        image[schema.AttributeItemGUID].String(),
		Published:   image[schema.AttributeItemPublishedAt].String(),
		Title:       image[schema.AttributeItemTitle].String(),
		Description: image[schema.AttributeItemDescription].String(),
		Link:        image[schema.AttributeItemLink].String(),
	}
}
