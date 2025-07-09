package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/tbtec/vdlg/internal/dto"
)

type S3Event struct {
	Records []struct {
		EventName string `json:"eventName"`
		S3        struct {
			Bucket struct {
				Name string `json:"name"`
			} `json:"bucket"`
			Object struct {
				Key string `json:"key"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}
type IConsumerService interface {
	ConsumeMessage(ctx context.Context) (*dto.Message, error)
}

type ConsumerService struct {
	QueueUrl string
	Client   *sqs.Client
	S3Client *s3.Client
}

func NewConsumerService(QueueUrl string, config aws.Config) IConsumerService {
	return &ConsumerService{
		QueueUrl: QueueUrl,
		Client:   sqs.NewFromConfig(config),
		S3Client: s3.NewFromConfig(config),
	}
}

func (consumer *ConsumerService) ConsumeMessage(ctx context.Context) (*dto.Message, error) {

	// Receive a message from the queue
	resp, err := consumer.Client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            &consumer.QueueUrl,
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     10,
	})

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	if len(resp.Messages) == 0 {
		return nil, nil // No messages available
	}

	fmt.Println("Raw message body:", *resp.Messages[0].Body)

	type SNSMessageWrapper struct {
		Message string `json:"Message"`
	}

	var snsMsg SNSMessageWrapper
	if err := json.Unmarshal([]byte(*resp.Messages[0].Body), &snsMsg); err != nil {
		fmt.Println("Erro ao fazer unmarshal do body SNS:", err)
		// Delete the message from the queue
		consumer.deleteSQSMessage(ctx, resp)
		return nil, err
	}

	var event S3Event
	if err := json.Unmarshal([]byte(snsMsg.Message), &event); err != nil {
		fmt.Println("Erro ao fazer unmarshal do campo Message:", err)
		// Delete the message from the queue
		consumer.deleteSQSMessage(ctx, resp)
		return nil, err
	}

	if len(event.Records) == 0 {
		fmt.Println("Nenhum registro encontrado no evento")
		// Delete the message from the queue
		consumer.deleteSQSMessage(ctx, resp)

		return nil, fmt.Errorf("evento sem registros")
	}

	record := event.Records[0]

	bucket := record.S3.Bucket.Name
	key := record.S3.Object.Key
	timestamp := time.Now().Format("20060102_150405")

	fmt.Printf("🎥 Novo vídeo: %s/%s - %s\n", bucket, key, timestamp)

	msg := dto.Message{
		BucketName: bucket,
		Key:        key,
	}

	slog.InfoContext(ctx, "Received message", "messageId", *resp.Messages[0].MessageId)
	slog.InfoContext(ctx, "Received message", "body", *resp.Messages[0].Body)

	// Delete the message from the queue
	consumer.deleteSQSMessage(ctx, resp)

	return &msg, nil
}

func (consumer *ConsumerService) deleteSQSMessage(ctx context.Context, resp *sqs.ReceiveMessageOutput) {
	out, delErr := consumer.Client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &consumer.QueueUrl,
		ReceiptHandle: resp.Messages[0].ReceiptHandle,
	})
	if delErr != nil {
		slog.ErrorContext(ctx, "Error deleting message", "error", delErr)
	}
	slog.InfoContext(ctx, "Message deleted", "recepit", *&out.ResultMetadata)
}

func (consumer *ConsumerService) DeleteMessage(ctx context.Context, receiptHandle string) error {
	_, err := consumer.Client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &consumer.QueueUrl,
		ReceiptHandle: &receiptHandle,
	})
	if err != nil {
		return err
	}

	return nil
}
