# lambda-feed-notifier
[![Build Status](https://travis-ci.com/pohzipohzi/lambda-feed-notifier.svg?branch=main)](https://travis-ci.com/pohzipohzi/lambda-feed-notifier)
[![Coverage Status](https://coveralls.io/repos/github/pohzipohzi/lambda-feed-notifier/badge.svg?branch=main)](https://coveralls.io/github/pohzipohzi/lambda-feed-notifier?branch=main)

Scrapes rss/atom feeds and notifies configured recipients of new items via AWS Lambda. The Lambda function is triggered by an EventBridge scheduled rule, and notifications are deduplicated by caching feed items on DynamoDB. A CloudFormation template is provided to provision the required resources.

## Cost

Costing depends on the time it takes to scrape and notify feeds, as well as the scraping interval. At the time of writing, scraping 5 feeds of around 10 items each and sending messages to a telegram bot every 5 minutes is within free tier.

## Usage

Configure Lambda environment variables in `template.yaml` (see comments in `Resources.LambdaFunction.Properties.Environment.Variables` for more information), then simply run:

```
./deploy <cloudformation-stack-name> <cloudformation-template-filepath>
```
