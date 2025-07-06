package usecase

import (
	"context"

	"github.com/tbtec/vdlg/internal/core/gateway"
	"github.com/tbtec/vdlg/internal/dto"
)

type UscProcessVideo struct {
	processorGateway *gateway.ProcessorGateway
	producerGateway  *gateway.ProducerGateway
}

func NewUscProcessVideo(processorGateway *gateway.ProcessorGateway, producerGateway *gateway.ProducerGateway) *UscProcessVideo {
	return &UscProcessVideo{
		processorGateway: processorGateway,
		producerGateway:  producerGateway,
	}
}

func (usc *UscProcessVideo) Process(ctx context.Context, url dto.Message) error {

	result := usc.processorGateway.ProcessVideo(ctx, url)

	err := usc.producerGateway.PublishMessage(ctx, result)

	if err != nil {
		return err
	}

	return nil
}
