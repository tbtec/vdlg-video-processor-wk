package usecase

import (
	"context"

	"github.com/tbtec/vdlg/internal/core/gateway"
	"github.com/tbtec/vdlg/internal/dto"
)

type UscProcessVideo struct {
	processorGateway *gateway.ProcessorGateway
}

func NewUscProcessVideo(processorGateway *gateway.ProcessorGateway) *UscProcessVideo {
	return &UscProcessVideo{
		processorGateway: processorGateway,
	}
}

func (usc *UscProcessVideo) Process(ctx context.Context, url dto.Message) error {

	err := usc.processorGateway.ProcessVideo(ctx, url)

	if err != nil {
		return err
	}

	return nil
}
