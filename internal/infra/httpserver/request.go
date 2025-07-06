package httpserver

import (
	"context"
	"encoding/json"
	"strconv"
)

// Request is the type tha represents a http request.
type Request struct {
	Host    string            `json:"host"`
	Path    string            `json:"path"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
	Params  map[string]string `json:"params"`
	Query   map[string]string `json:"queryParams"`
}

// ParseQuery parses the query string parameter with the given name
func (req Request) ParseQuery(name string) string {
	return req.Query[name]
}

// ParseQueryInt parses the query string parameter with the given name
func (req Request) ParseQueryInt(name string) int {
	intVar, err := strconv.Atoi(req.Query[name])
	if err != nil {
		return 0
	}
	return intVar
}

// ParseParam parses the path parameter with the given name
func (req Request) ParseParamString(name string) string {
	return req.Params[name]
}

// ParseParamInt parses the path parameter with the given name
func (req Request) ParseParamInt(name string) int {
	intVar, err := strconv.Atoi(req.Params[name])
	if err != nil {
		return 0
	}
	return intVar
}

// ParseHeader parses the header with the given name
func (req Request) ParseHeader(name string) string {
	return req.Headers[name]
}

// ParseHeaderInt parses the header with the given name
func (req Request) ParseHeaderInt(name string) int {
	intVar, err := strconv.Atoi(req.Headers[name])
	if err != nil {
		return 0
	}
	return intVar
}

// ParseBody parses the request body into the given result
func (req Request) ParseBody(ctx context.Context, result any) error {
	err := json.Unmarshal(req.Body, result)
	if err != nil {
		return err
	}
	return nil
}
