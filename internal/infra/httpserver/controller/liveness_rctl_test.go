package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tbtec/vdlg/internal/infra/httpserver"
)

func TestLivenessControllerHandle(t *testing.T) {
	ctrl := NewLivenessController()
	req := httpserver.Request{} // pode ser vazio para este teste

	resp := ctrl.Handle(context.Background(), req)

	assert.Equal(t, 200, resp.Code)
	output, ok := resp.Body.(Output)
	assert.True(t, ok)
	assert.Equal(t, "OK", output.Status)
}
