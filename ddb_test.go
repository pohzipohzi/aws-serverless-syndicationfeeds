package main

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ddbTestEndpoint        = "http://localhost:8000"
	ddbTestAccessKeyID     = "fakeMyKeyId"
	ddbTestSecretAccessKey = "fakeSecretAccessKey"
)

func Test_DdbUpdateAndGet(t *testing.T) {
	sess, err := session.NewSession(
		aws.NewConfig().
			WithEndpoint(ddbTestEndpoint).
			WithCredentials(credentials.NewStaticCredentials(ddbTestAccessKeyID, ddbTestSecretAccessKey, "")),
	)
	require.Nil(t, err)
	ddb := dynamodb.New(sess)
	_, err = ddb.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(ddbAttributeTitle),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String(ddbAttributeItemGUID),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(ddbAttributeTitle),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String(ddbAttributeItemGUID),
				KeyType:       aws.String("RANGE"),
			},
		},
		TableName:   aws.String(ddbTableFeed),
		BillingMode: aws.String("PAY_PER_REQUEST"),
	})
	require.Nil(t, err)
	defer func() {
		_, err = ddb.DeleteTable(&dynamodb.DeleteTableInput{
			TableName: aws.String(ddbTableFeed),
		})
		require.Nil(t, err)
	}()
	feeds := []*gofeed.Feed{
		{
			Title: "feed 1",
			Items: []*gofeed.Item{
				{
					GUID:            "guid1",
					Title:           "item title 1",
					Description:     "item description 1",
					Link:            "link 1",
					PublishedParsed: pt(time.Unix(1, 0)),
				},
				{
					GUID:            "guid2",
					Title:           "item title 2",
					Description:     "item description 2",
					Link:            "link 2",
					PublishedParsed: pt(time.Unix(2, 0)),
				},
			},
		},
		{
			Title: "feed 2",
			Items: []*gofeed.Item{
				{
					GUID:            "guid1",
					Title:           "item title 1",
					Description:     "item description 1",
					Link:            "link 1",
					PublishedParsed: pt(time.Unix(3, 0)),
				},
			},
		},
	}
	for _, f := range feeds {
		for _, i := range f.Items {
			err := ddbUpdateItem(ddb, f, i)
			require.Nil(t, err)
		}
	}
	for _, f := range feeds {
		res, err := ddbGetItems(ddb, f)
		require.Nil(t, err)
		for _, expect := range f.Items {
			assert.Equal(t, expect, res[FeedCompositeKey{
				Title:    f.Title,
				ItemGUID: expect.GUID,
			}])
		}
	}
}

func Test_TimeConversion(t *testing.T) {
	tm := time.Unix(1, 0)
	assert.Equal(t, &tm, atot(ttoa(&tm)))
}

func Test_shouldHandleItem(t *testing.T) {
	for _, test := range []struct {
		desc   string
		f      *gofeed.Feed
		i      *gofeed.Item
		di     map[FeedCompositeKey]*gofeed.Item
		expect bool
	}{
		{
			desc:   "feed is nil",
			f:      nil,
			i:      nil,
			di:     nil,
			expect: false,
		},
		{
			desc:   "item is nil",
			f:      &gofeed.Feed{},
			i:      nil,
			di:     nil,
			expect: false,
		},
		{
			desc:   "feed and item are non-nil, but ddb items are nil",
			f:      &gofeed.Feed{},
			i:      &gofeed.Item{},
			di:     nil,
			expect: true,
		},
		{
			desc:   "item does not exist in ddb",
			f:      &gofeed.Feed{Title: "t"},
			i:      &gofeed.Item{GUID: "g"},
			di:     map[FeedCompositeKey]*gofeed.Item{},
			expect: true,
		},
		{
			desc: "item exists in ddb",
			f:    &gofeed.Feed{Title: "t"},
			i:    &gofeed.Item{GUID: "g", PublishedParsed: pt(time.Unix(1, 0))},
			di: map[FeedCompositeKey]*gofeed.Item{
				{Title: "t", ItemGUID: "g"}: {PublishedParsed: pt(time.Unix(1, 0))},
			},
			expect: false,
		},
		{
			desc: "item exists in ddb, but pubdate is greater",
			f:    &gofeed.Feed{Title: "t"},
			i:    &gofeed.Item{GUID: "g", PublishedParsed: pt(time.Unix(2, 0))},
			di: map[FeedCompositeKey]*gofeed.Item{
				{Title: "t", ItemGUID: "g"}: {PublishedParsed: pt(time.Unix(1, 0))},
			},
			expect: true,
		},
	} {
		assert.Equal(t, test.expect, shouldHandleItem(test.f, test.i, test.di), test.desc)
	}
}

func pt(in time.Time) *time.Time {
	return &in
}
