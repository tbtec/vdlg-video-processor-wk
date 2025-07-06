package httpclient

import (
	"github.com/go-resty/resty/v2"
)

type Client = resty.Client

func New() *Client {
	return resty.New()
}
