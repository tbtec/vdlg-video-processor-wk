package httpserver

import (
	"context"

	"github.com/tbtec/vdlg/internal/types/xerrors"
)

// HandleError converts error to response
func HandleError(ctx context.Context, err error) Response {
	switch codError := err.(type) {
	case xerrors.ValidationError:
		var valErrs []DetailResponse
		for _, f := range codError.Fields {
			valErrs = append(valErrs, DetailResponse{f.Name, f.Reasons})
		}
		return BadRequest(NewErrorMessage("400", "Bad Request", valErrs...))

	case xerrors.BusinessError:
		return UnprocessableEntity(NewErrorMessage(codError.Code, codError.Description))
	case xerrors.NotFoundError:
		return NotFound(NewErrorMessage("404", codError.Description))
	default:
		return InternalServerError(NewErrorMessage("500", "Internal Server Error"))
	}
}
