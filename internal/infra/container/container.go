package container

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/tbtec/vdlg/internal/env"

	"github.com/tbtec/vdlg/internal/infra/event"
)

type Container struct {
	Config          env.Config
	ConsumerService event.IConsumerService
	ProducerService event.IProducerService
	AwsConfig       aws.Config
}

func New(config env.Config) (*Container, error) {
	factory := Container{}
	factory.Config = config

	return &factory, nil
}

func (container *Container) Start(ctx context.Context) error {

	var err error

	if container.Config.Env == "local-stack" { // LocalStack
		//container.AwsConfig = container.GetLocalStackConfig(ctx)
	} else {
		container.AwsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(container.Config.AwsRegion))
		if err != nil {
			log.Fatalf("erro ao carregar config: %v", err)
		}
	}

	container.ConsumerService = event.NewConsumerService(container.Config.InputQueueUrl, container.AwsConfig)
	container.ProducerService = event.NewProducerService(container.Config.ProcessResultTopicArn, container.AwsConfig)

	return nil
}

/*func (container *Container) GetLocalStackConfig(ctx context.Context) aws.Config {

	awsConfig, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	awsConfig.BaseEndpoint = aws.String("http://localhost:4566")

	if err != nil {
		log.Fatalf("erro ao carregar config: %v", err)
	}

	return awsConfig
}*/
