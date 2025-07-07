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
		return nil, err
	}

	var event S3Event
	if err := json.Unmarshal([]byte(snsMsg.Message), &event); err != nil {
		fmt.Println("Erro ao fazer unmarshal do campo Message:", err)

		out, delErr := consumer.Client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      &consumer.QueueUrl,
			ReceiptHandle: resp.Messages[0].ReceiptHandle,
		})
		if delErr != nil {
			slog.ErrorContext(ctx, "Error deleting message", "error", delErr)
		}
		slog.InfoContext(ctx, "Message deleted", "recepit", *&out.ResultMetadata)

		return nil, err
	}

	if len(event.Records) == 0 {
		fmt.Println("Nenhum registro encontrado no evento")
		return nil, fmt.Errorf("evento sem registros")
	}

	record := event.Records[0]

	bucket := record.S3.Bucket.Name
	key := record.S3.Object.Key
	timestamp := time.Now().Format("20060102_150405")

	fmt.Printf("🎥 Novo vídeo: %s/%s - %s\n", bucket, key, timestamp)

	url, err := consumer.generatePresignedURLV2(ctx, bucket, key)
	if err != nil {
		fmt.Println("Erro ao gerar URL assinada:", err)
		//continue
	}

	msg := dto.Message{
		BucketName: bucket,
		Key:        key,
		Url:        url,
	}

	slog.InfoContext(ctx, "Received message", "messageId", *resp.Messages[0].MessageId)
	slog.InfoContext(ctx, "Received message", "body", *resp.Messages[0].Body)
	slog.InfoContext(ctx, "Received message", "presignedURL", url)

	// Delete the message from the queue
	out, delErr := consumer.Client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &consumer.QueueUrl,
		ReceiptHandle: resp.Messages[0].ReceiptHandle,
	})
	if delErr != nil {
		slog.ErrorContext(ctx, "Error deleting message", "error", delErr)
	}
	slog.InfoContext(ctx, "Message deleted", "recepit", *&out.ResultMetadata)

	return &msg, nil
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

func (consumer *ConsumerService) generatePresignedURLV2(ctx context.Context, bucket, key string) (string, error) {
	presignClient := s3.NewPresignClient(consumer.S3Client)

	presignedURL, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(15*time.Minute))

	if err != nil {
		return "", fmt.Errorf("failed to sign request: %w", err)
	}

	return presignedURL.URL, nil
}

/*func (consumer *ConsumerService) GetObjectFromS3(ctx context.Context, bucket, key string) ([]byte, error) {
	getObjInput := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	resp, err := consumer.S3Client.GetObject(ctx, getObjInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	return data, nil
}*/
