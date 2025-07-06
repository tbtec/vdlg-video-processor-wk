package event

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type IProducerService interface {
	PublishMessage(ctx context.Context, message interface{}) error
}

type ProducerService struct {
	TopicName string
	TopicArn  string
	Client    *sns.Client
}

func NewProducerService(topicArn string, config aws.Config) IProducerService {
	return &ProducerService{
		TopicArn: topicArn,
		Client:   sns.NewFromConfig(config),
	}
}
func (producer *ProducerService) PublishMessage(ctx context.Context, message interface{}) error {
	// Serialize order to JSON
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	input := &sns.PublishInput{
		TopicArn: aws.String(producer.TopicArn),
		Message:  aws.String(string(body)),
	}

	slog.InfoContext(ctx, "Publishing message", "order", string(body))
	slog.InfoContext(ctx, "Publishing message", "topicArn", producer.TopicArn)

	output, err := producer.Client.Publish(ctx, input)

	slog.InfoContext(ctx, "Message published", "recepit", output.MessageId)
	if err != nil {
		return err
	}

	return nil
}
