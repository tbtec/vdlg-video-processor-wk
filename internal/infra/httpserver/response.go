package httpserver

// Response is the type tha represents a http response.
type Response struct {
	Code    int
	Body    any
	Headers map[string]string
}

// Ok returns a 200 response with the given body.
func Ok(body any) Response {
	return Response{200, body, nil}
}

// Created returns a 200 response with the given body.
func Created(body any) Response {
	return Response{201, body, nil}
}

// Accepted returns a 202 response with the given body.
func Accepted(body any) Response {
	return Response{202, body, nil}
}

// NoContent returns a 204 response without a body.
func NoContent() Response {
	return Response{204, nil, nil}
}

// BadRequest returns a 400 response.
func BadRequest(body any) Response {
	return Response{400, body, nil}
}

// NotFound returns a 404 response with a body
func NotFound(body any) Response {
	return Response{404, body, nil}
}

// Conflict returns a 409 with a body
func Conflict(body any) Response {
	return Response{409, body, nil}
}

// UnprocessableEntity returns a 422 response with a body
func UnprocessableEntity(body any) Response {
	return Response{422, body, nil}
}

// InternalServerError returns a 500 response with the given error.
func InternalServerError(body any) Response {
	return Response{500, body, nil}
}

// ServiceUnavailable returns a 503 response with a body
func ServiceUnavailable(body any) Response {
	return Response{503, body, nil}
}
