package env

import (
	"context"
	"log/slog"
	"os"

	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
)

type Config struct {
	Env                   string `env:"ENV" envDefault:"local"`
	Port                  int    `env:"PORT" envDefault:"8080"`
	InputQueueUrl         string `env:"INPUT_QUEUE_URL"`
	ProcessResultTopicArn string `env:"PROCESS_RESULT_TOPIC_ARN"`
	AwsAccessKeyId        string `env:"AWS_ACCESS_KEY_ID"`
	AwsSecretAccessKey    string `env:"AWS_SECRET_ACCESS_KEY"`
	AwsSessionToken       string `env:"AWS_SESSION_TOKEN"`
	AwsRegion             string `env:"AWS_REGION"`
	AwsUseCredentials     string `env:"AWS_USE_CREDENTIALS" envDefault:"false"`
}

func LoadEnvConfig() (Config, error) {
	cfg := Config{}
	var err error

	if fileExists("../../.env") {
		err = godotenv.Load("../../.env")
	}
	if fileExists("../.env") {
		err = godotenv.Load("../.env")
	}
	if fileExists(".env") {
		err = godotenv.Load(".env")
	}
	if err != nil {
		slog.ErrorContext(context.Background(), "❌ Could not load .env file")
		panic(err)
	}
	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}
	slog.InfoContext(context.Background(), "✅ Success on read .env configuration...")

	return cfg, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return info != nil
}
