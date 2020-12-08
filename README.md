# AWS Serverless Syndication Feeds
[![Build Status](https://travis-ci.com/pohzipohzi/aws-serverless-syndicationfeeds.svg?branch=main)](https://travis-ci.com/pohzipohzi/aws-serverless-syndicationfeeds)
[![Coverage Status](https://coveralls.io/repos/github/pohzipohzi/aws-serverless-syndicationfeeds/badge.svg?branch=main)](https://coveralls.io/github/pohzipohzi/aws-serverless-syndicationfeeds?branch=main)

This is a simple AWS-based system that scrapes syndication feeds and writes feed items into a DynamoDB table via Lambda. Scheduled rules can be set up on EventBridge to automatically trigger the Lambda function for each target URL. DynamoDB streams can be used in conjunction with other AWS services to do useful work on new feed items. A CloudFormation template is provided to provision the required resources.

<p align="center">
  <img src="arch.svg">
</p>

## Cost

At the time of writing, scraping 5 feeds every 5 minutes is safely within free tier. The author has no experience with more expensive setups.

## Deployment

This project leverages AWS's [Serverless Application Model (SAM)](https://docs.aws.amazon.com/serverless-application-model/) for deployment.

First configure `template.yaml`, then simply run:

```
CGO_ENABLED=0 sam build
sam deploy --guided
```
