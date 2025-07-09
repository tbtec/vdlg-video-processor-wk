package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tbtec/vdlg/internal/env"
	"github.com/tbtec/vdlg/internal/infra/container"
	"github.com/tbtec/vdlg/internal/infra/event/eventserver"
	"github.com/tbtec/vdlg/internal/infra/httpserver/server"
)

func main() {

	ctx := context.Background()

	if err := run(ctx); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {

	slog.InfoContext(ctx, "LoadEnvConfig...")
	config, err := env.LoadEnvConfig()
	if err != nil {
		log.Fatal(err)
	}

	slog.InfoContext(ctx, "New Container...")
	container, err := container.New(config)
	if err != nil {
		log.Fatal(err)
	}

	slog.InfoContext(ctx, "Starting Container...")
	errStart := container.Start(ctx)
	if errStart != nil {
		log.Fatal(err)
	}

	slog.InfoContext(ctx, "New Server...")
	httpServer := server.New(container, config)

	slog.InfoContext(ctx, "New Event Server...")
	eventServer := eventserver.NewEventServer(container, config)

	slog.InfoContext(ctx, "Starting Event Server...")

	go func(ctx context.Context) {
		for {
			eventServer.Consume(ctx)
		}
	}(ctx)

	httpServer.Listen()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sc

	ctx, shutdown := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdown()

	slog.InfoContext(ctx, "Shutting down services...")

	return nil
}
