package adapter

import (
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mmcdole/gofeed"
)

const (
	ddbAttributeTitle           = "Title"
	ddbAttributeItemGUID        = "ItemGUID"
	ddbAttributeItemPublishedAt = "ItemPublishedAt"
	ddbAttributeItemTitle       = "ItemTitle"
	ddbAttributeItemDescription = "ItemDescription"
	ddbAttributeItemLink        = "ItemLink"
)

func FeedItemToDdbUpdateItemInput(f *gofeed.Feed, i *gofeed.Item, table string) *dynamodb.UpdateItemInput {
	if f == nil || i == nil {
		return nil
	}
	return &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#P": aws.String(ddbAttributeItemPublishedAt),
			"#T": aws.String(ddbAttributeItemTitle),
			"#D": aws.String(ddbAttributeItemDescription),
			"#L": aws.String(ddbAttributeItemLink),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":p": {S: aws.String(i.PublishedParsed.Format(time.RFC3339))},
			":t": {S: aws.String(i.Title)},
			":d": {S: aws.String(i.Description)},
			":l": {S: aws.String(i.Link)},
		},
		Key: map[string]*dynamodb.AttributeValue{
			ddbAttributeTitle:    {S: aws.String(f.Title)},
			ddbAttributeItemGUID: {S: aws.String(i.GUID)},
		},
		TableName:        aws.String(table),
		UpdateExpression: aws.String("SET #P = :p, #T = :t, #D = :d, #L = :l"),
	}
}

func DdbImageToFeedItem(image map[string]events.DynamoDBAttributeValue) *gofeed.Item {
	if image == nil {
		return nil
	}
	return &gofeed.Item{
		GUID:        image[ddbAttributeItemGUID].String(),
		Published:   image[ddbAttributeItemPublishedAt].String(),
		Title:       image[ddbAttributeItemTitle].String(),
		Description: image[ddbAttributeItemDescription].String(),
		Link:        image[ddbAttributeItemLink].String(),
	}
}
