package httpserver

import "context"

type IController interface {
	Handle(ctx context.Context, request Request) Response
}
