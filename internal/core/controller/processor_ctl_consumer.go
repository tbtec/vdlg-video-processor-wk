package controller

import (
	"context"

	"github.com/tbtec/vdlg/internal/core/domain/usecase"
	"github.com/tbtec/vdlg/internal/core/gateway"
	"github.com/tbtec/vdlg/internal/dto"
	"github.com/tbtec/vdlg/internal/infra/container"
)

type ConsumerController struct {
	usc *usecase.UscProcessVideo
}

func NewConsumerController(container *container.Container) *ConsumerController {
	return &ConsumerController{
		usc: usecase.NewUscProcessVideo(
			gateway.NewProcessorGateway(container.AwsConfig),
			gateway.NewProducerGateway(container.ProducerService),
		),
	}
}

func (ctl *ConsumerController) Execute(ctx context.Context, url dto.Message) error {
	return ctl.usc.Process(ctx, url)
}
