package eventserver

import (
	"context"
	"log/slog"

	"github.com/tbtec/vdlg/internal/core/controller"
	"github.com/tbtec/vdlg/internal/env"
	"github.com/tbtec/vdlg/internal/infra/container"
	"github.com/tbtec/vdlg/internal/infra/event"
)

type EventServer struct {
	ConsumerService    event.IConsumerService
	ConsumerController *controller.ConsumerController
}

func NewEventServer(container *container.Container, config env.Config) *EventServer {
	slog.InfoContext(context.Background(), "Creating Event Server...")

	cpc := controller.NewConsumerController(container)
	cs := container.ConsumerService

	return &EventServer{
		ConsumerService:    cs,
		ConsumerController: cpc,
	}

}

func (eventServer *EventServer) Consume(ctx context.Context) {

	// Start the consumer service
	url, err := eventServer.ConsumerService.ConsumeMessage(ctx)

	if err != nil {
		slog.ErrorContext(ctx, "Error reading message ", err)
	}
	if url == nil {
		slog.InfoContext(ctx, "No messages available")
	} else {
		slog.InfoContext(ctx, "Processing message...")
		err2 := eventServer.ConsumerController.Execute(ctx, *url)
		if err2 != nil {
			slog.ErrorContext(ctx, "Error processing message: ", err2)

		}
	}
}
