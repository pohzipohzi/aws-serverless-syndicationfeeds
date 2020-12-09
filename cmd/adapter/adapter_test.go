package adapter

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
)

func Test_FeedItemToDdbUpdateItemInput(t *testing.T) {
	for _, test := range []struct {
		f      *gofeed.Feed
		i      *gofeed.Item
		table  string
		expect *dynamodb.UpdateItemInput
	}{
		{
			f:      nil,
			i:      nil,
			table:  "",
			expect: nil,
		},
	} {
		assert.Equal(t, test.expect, FeedItemToDdbUpdateItemInput(test.f, test.i, test.table))
	}
}

func Test_DdbImageToFeedItem(t *testing.T) {
	for _, test := range []struct {
		image  map[string]events.DynamoDBAttributeValue
		expect *gofeed.Item
	}{
		{
			image:  nil,
			expect: nil,
		},
	} {
		assert.Equal(t, test.expect, DdbImageToFeedItem(test.image))
	}
}
