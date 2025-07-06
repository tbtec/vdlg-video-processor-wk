package xerrors

type BusinessError struct {
	Description string
	Code        string
}

func (e BusinessError) Error() string {
	return e.Code + " - " + e.Description
}

func NewBusinessError(code string, desc string) BusinessError {
	return BusinessError{
		Code:        code,
		Description: desc,
	}
}
