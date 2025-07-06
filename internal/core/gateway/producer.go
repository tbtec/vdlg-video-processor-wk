package gateway

import (
	"context"

	"github.com/tbtec/vdlg/internal/infra/event"
)

type ProducerGateway struct {
	producerService event.IProducerService
}

func NewProducerGateway(producerService event.IProducerService) *ProducerGateway {
	return &ProducerGateway{
		producerService: producerService,
	}
}

func (gtw *ProducerGateway) PublishMessage(ctx context.Context, result ProcessingResult) error {

	err := gtw.producerService.PublishMessage(ctx, result)
	if err != nil {
		return err
	}

	return nil
}
