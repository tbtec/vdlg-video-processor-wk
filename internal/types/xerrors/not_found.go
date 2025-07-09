package xerrors

type NotFoundError struct {
	Description string
	Code        string
}

func (e NotFoundError) Error() string {
	return "Not Found - " + e.Code + e.Description
}

func NewNotFoundError(code string, desc string) NotFoundError {
	return NotFoundError{
		Description: desc,
		Code:        code,
	}
}
