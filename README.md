# lambda-feed-notifier

## Overview

Scrapes rss/atom feeds for new items and notifies configured recipients via AWS Lambda. The Lambda function is triggered by an EventBridge scheduled rule, and events are deduplicated by caching them on DynamoDB. A CloudFormation template is provided to provision the required resources.

## Cost

Costing depends on the number of feeds scraped and the scraping interval. At the time of writing, scraping <10 feeds every 5 minutes lands me safely within free tier.

## Usage

Configure Lambda environment variables in `template.yaml` (see comments in `Resources.LambdaFunction.Properties.Environment.Variables` for more information), then simply run:

```
./deploy <cloudformation-stack-name> <cloudformation-template-filepath>
```
