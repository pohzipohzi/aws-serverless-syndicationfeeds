package main

import (
	"lambda-feed-notifier/internal"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(internal.Main)
}
