package main

import (
	schema "aws-serverless-syndicationfeeds/cmd/ddb_schema"
	"errors"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/go-multierror"
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
	log.Printf("retrieved %d items\n", len(feed.Items))
	for _, i := range feed.Items {
		err = writeFeedItem(ddb, ddbTableName, feed, i)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs.ErrorOrNil()
}

func writeFeedItem(ddb *dynamodb.DynamoDB, table string, f *gofeed.Feed, i *gofeed.Item) error {
	if ddb == nil || f == nil {
		return nil
	}

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#P": aws.String(schema.AttributeItemPublishedAt),
			"#T": aws.String(schema.AttributeItemTitle),
			"#D": aws.String(schema.AttributeItemDescription),
			"#L": aws.String(schema.AttributeItemLink),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":p": {S: aws.String(i.PublishedParsed.Format(time.RFC3339))},
			":t": {S: aws.String(i.Title)},
			":d": {S: aws.String(i.Description)},
			":l": {S: aws.String(i.Link)},
		},
		Key: map[string]*dynamodb.AttributeValue{
			schema.AttributeTitle:    {S: aws.String(f.Title)},
			schema.AttributeItemGUID: {S: aws.String(i.GUID)},
		},
		TableName:        aws.String(table),
		UpdateExpression: aws.String("SET #P = :p, #T = :t, #D = :d, #L = :l"),
	}
	_, err := ddb.UpdateItem(input)
	return err
}
