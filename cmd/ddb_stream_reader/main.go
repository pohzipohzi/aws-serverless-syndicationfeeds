package main

import (
	schema "lambda-feed-notifier/cmd/ddb_schema"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	lambda.Start(handler)
}

func handler(e events.DynamoDBEvent) error {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	for _, r := range e.Records {
		item := feedItemFromImage(r.Change.NewImage)
		log.Info().Msgf("Received feed item: %v\n", item)
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
