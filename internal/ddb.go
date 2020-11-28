package internal

import (
	"time"

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

type FeedCompositeKey struct {
	Title    string
	ItemGUID string
}

func ddbGetItems(ddb *dynamodb.DynamoDB, f *gofeed.Feed) (map[FeedCompositeKey]*gofeed.Item, error) {
	if ddb == nil || f == nil {
		return nil, nil
	}
	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			envDdbTableName: {Keys: keysFromFeed(f)},
		},
	}
	res, err := ddb.BatchGetItem(input)
	if err != nil {
		return nil, err
	}
	return itemMapFromBatchGetItemOutput(res), nil
}

func keysFromFeed(f *gofeed.Feed) []map[string]*dynamodb.AttributeValue {
	ret := []map[string]*dynamodb.AttributeValue{}
	for _, i := range f.Items {
		if i == nil {
			continue
		}
		ret = append(ret, map[string]*dynamodb.AttributeValue{
			ddbAttributeTitle:    {S: aws.String(f.Title)},
			ddbAttributeItemGUID: {S: aws.String(i.GUID)},
		})
	}
	return ret
}

func itemMapFromBatchGetItemOutput(res *dynamodb.BatchGetItemOutput) map[FeedCompositeKey]*gofeed.Item {
	if res == nil || res.Responses == nil || res.Responses[envDdbTableName] == nil {
		return nil
	}
	ret := map[FeedCompositeKey]*gofeed.Item{}
	for _, r := range res.Responses[envDdbTableName] {
		if r == nil {
			continue
		}
		ret[FeedCompositeKey{
			Title:    aws.StringValue(r[ddbAttributeTitle].S),
			ItemGUID: aws.StringValue(r[ddbAttributeItemGUID].S),
		}] = &gofeed.Item{
			GUID:            aws.StringValue(r[ddbAttributeItemGUID].S),
			PublishedParsed: atot(aws.StringValue(r[ddbAttributeItemPublishedAt].S)),
			Title:           aws.StringValue(r[ddbAttributeItemTitle].S),
			Description:     aws.StringValue(r[ddbAttributeItemDescription].S),
			Link:            aws.StringValue(r[ddbAttributeItemLink].S),
		}
	}
	return ret
}

func ddbUpdateItem(ddb *dynamodb.DynamoDB, f *gofeed.Feed, i *gofeed.Item) error {
	if ddb == nil || f == nil || i == nil {
		return nil
	}
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#P": aws.String(ddbAttributeItemPublishedAt),
			"#T": aws.String(ddbAttributeItemTitle),
			"#D": aws.String(ddbAttributeItemDescription),
			"#L": aws.String(ddbAttributeItemLink),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":p": {S: aws.String(ttoa(i.PublishedParsed))},
			":t": {S: aws.String(i.Title)},
			":d": {S: aws.String(i.Description)},
			":l": {S: aws.String(i.Link)},
		},
		Key: map[string]*dynamodb.AttributeValue{
			ddbAttributeTitle:    {S: aws.String(f.Title)},
			ddbAttributeItemGUID: {S: aws.String(i.GUID)},
		},
		TableName:        aws.String(envDdbTableName),
		UpdateExpression: aws.String("SET #P = :p, #T = :t, #D = :d, #L = :l"),
	}
	_, err := ddb.UpdateItem(input)
	return err
}

func ttoa(t *time.Time) string {
	return t.Format(time.RFC3339)
}

func atot(a string) *time.Time {
	t, _ := time.Parse(time.RFC3339, a)
	return &t
}

func shouldHandleItem(f *gofeed.Feed, i *gofeed.Item, di map[FeedCompositeKey]*gofeed.Item) bool {
	if f == nil || i == nil {
		return false
	}
	if di == nil {
		return true
	}
	if v, ok := di[FeedCompositeKey{
		Title:    f.Title,
		ItemGUID: i.GUID,
	}]; ok {
		if v == nil || i.PublishedParsed == nil || v.PublishedParsed == nil {
			return true
		}
		return i.PublishedParsed.After(*v.PublishedParsed)
	}
	return true
}
