package xerrors

import "fmt"

const (
	ReasonTypeInvalidValue         = "INVALID_VALUE"
	ReasonRequiredAttributeMissing = "REQUIRED_ATTRIBUTE_MISSING"
)

type ValidationError struct {
	Description string
	Fields      []Field
}

type Field struct {
	Name    string
	Reasons []string
}

func NewValidationError(desc string) ValidationError {
	return ValidationError{
		Description: desc,
		Fields:      []Field{},
	}
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%d fields are invalid", len(e.Fields))
}

func (e ValidationError) AddField(attr string, reasons ...string) ValidationError {
	e.Fields = append(e.Fields, Field{attr, reasons})
	return e
}
