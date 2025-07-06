package httpserver

type ErrorMessage struct {
	Error Error `json:"error"`
}
type Error struct {
	Description string           `json:"description"`
	Code        string           `json:"code,omitempty"`
	Details     []DetailResponse `json:"details,omitempty"`
}

type DetailResponse struct {
	Attribute string   `json:"attribute"`
	Messages  []string `json:"messages"`
}

func NewErrorMessage(code string, desc string, details ...DetailResponse) ErrorMessage {
	return ErrorMessage{
		Error: Error{
			Code:        code,
			Description: desc,
			Details:     details,
		},
	}
}
