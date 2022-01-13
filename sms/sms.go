// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX - License - Identifier: Apache - 2.0
package sms

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// SNSPublishAPI defines the interface for the Publish function.
// We use this interface to test the function using a mocked service.
type SNSPublishAPI interface {
	Publish(ctx context.Context,
		params *sns.PublishInput,
		optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

// PublishMessage publishes a message to an Amazon Simple Notification Service (Amazon SNS) topic
// Inputs:
//     c is the context of the method call, which includes the Region
//     api is the interface that defines the method call
//     input defines the input arguments to the service call.
// Output:
//     If success, a PublishOutput object containing the result of the service call and nil
//     Otherwise, nil and an error from the call to Publish
func PublishMessage(c context.Context, api SNSPublishAPI, input *sns.PublishInput) (*sns.PublishOutput, error) {
	return api.Publish(c, input)
}

func SendTextPhone(msg string, phone string) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	cfg.Region = "us-east-1"

	client := sns.NewFromConfig(cfg)

	input := &sns.PublishInput{
		Message:     &msg,
		PhoneNumber: &phone,
	}

	result, err := PublishMessage(context.TODO(), client, input)
	if err != nil {
		fmt.Println("Got an error publishing the message:")
		fmt.Println(err)
		return
	}

	fmt.Println("Message ID: " + *result.MessageId)
}

func SendTextTopic(msg string, topicARN string) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	cfg.Region = "us-east-1"

	client := sns.NewFromConfig(cfg)

	input := &sns.PublishInput{
		Message:  &msg,
		TopicArn: &topicARN,
	}

	result, err := PublishMessage(context.TODO(), client, input)
	if err != nil {
		fmt.Println("Got an error publishing the message:")
		fmt.Println(err)
		return
	}

	fmt.Println("Message ID: " + *result.MessageId)
}
