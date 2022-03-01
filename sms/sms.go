// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX - License - Identifier: Apache - 2.0
package sms

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go/aws"
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

func SendTextPhone(msg string, phone string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}
	cfg.Region = "us-east-1"

	client := sns.NewFromConfig(cfg)

	input := &sns.PublishInput{
		Message:     &msg,
		PhoneNumber: &phone,
	}

	result, err := PublishMessage(context.TODO(), client, input)
	if err != nil {
		return "", err
	}

	log.Println("Message ID: " + *result.MessageId)

	return *result.MessageId, nil
}

func SendTextTopic(msg string, topicARN string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}
	cfg.Region = "us-east-1"

	client := sns.NewFromConfig(cfg)

	input := &sns.PublishInput{
		Message:  &msg,
		TopicArn: &topicARN,
	}

	result, err := PublishMessage(context.TODO(), client, input)
	if err != nil {
		return "", err
	}

	log.Println("Message ID: " + *result.MessageId)

	return *result.MessageId, nil
}

func SubscribeUser(phone string, topicARN string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}
	cfg.Region = "us-east-1"

	client := sns.NewFromConfig(cfg)

	input := &sns.SubscribeInput{
		Endpoint:              &phone,
		Protocol:              aws.String("sms"),
		ReturnSubscriptionArn: true,
		TopicArn:              aws.String(topicARN),
	}

	result, err := client.Subscribe(context.TODO(), input)
	if err != nil {
		return "", err
	}

	return *result.SubscriptionArn, nil
}

func RemoveSubscriber(subscriptionArn string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}
	cfg.Region = "us-east-1"

	client := sns.NewFromConfig(cfg)

	input := &sns.UnsubscribeInput{
		SubscriptionArn: &subscriptionArn,
	}

	_, err = client.Unsubscribe(context.TODO(), input)
	if err != nil {
		return err
	}

	return nil
}

func ConfirmSubscribeUser(phone string, topicARN string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}
	cfg.Region = "us-east-1"

	client := sns.NewFromConfig(cfg)

	input := &sns.SubscribeInput{
		Endpoint:              &phone,
		Protocol:              aws.String("sms"),
		ReturnSubscriptionArn: true,
		TopicArn:              aws.String(topicARN),
	}

	result, err := client.Subscribe(context.TODO(), input)
	if err != nil {
		return err
	}

	log.Println("Subscription ARN: " + *result.SubscriptionArn)

	return nil
}

func CreateTopic(topic string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}
	cfg.Region = "us-east-1"

	client := sns.NewFromConfig(cfg)

	input := &sns.CreateTopicInput{
		Name:       &topic,
		Attributes: map[string]string{"DisplayName": topic},
	}

	result, err := client.CreateTopic(context.TODO(), input)
	if err != nil {
		return "", err
	}

	return *result.TopicArn, nil
}

func DeleteTopic(topicARN string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}
	cfg.Region = "us-east-1"

	client := sns.NewFromConfig(cfg)

	input := &sns.DeleteTopicInput{
		TopicArn: &topicARN,
	}

	_, err = client.DeleteTopic(context.TODO(), input)
	if err != nil {
		return err
	}

	return nil
}
